package guardian

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	goVersionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+(?:\.[0-9]+)?$`)
	imagePattern     = regexp.MustCompile(`^hermes-guardian:[0-9]+\.[0-9]+(?:\.[0-9]+)?$`)
)

var generatedPaths = map[string]struct{}{
	"spec/bot-api.json":                  {},
	"spec/bot-api-semantics.json":        {},
	"spec/api-surface.txt":               {},
	"types/zz_botapi_generated.go":       {},
	"api/zz_botapi_aliases_generated.go": {},
	"zz_botapi_aliases_generated.go":     {},
	".github/guardian/update.json":       {},
	"CHANGELOG.md":                       {},
	"README.md":                          {},
	"docs/schema-parity.md":              {},
	"api/client.go":                      {},
}

// Controller owns one repository-scoped Guardian run.
type Controller struct {
	Root string
	Out  string
	Work string
	Env  func(string) string
	API  *GitHubClient
}

func NewController(root string) (*Controller, error) {
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(absolute)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("guardian: repository root is unavailable")
	}
	out := filepath.Join(absolute, ".guardian")
	return &Controller{
		Root: absolute,
		Out:  out,
		Work: filepath.Join(out, "work"),
		Env:  os.Getenv,
	}, nil
}

func (controller *Controller) getenv(key string) string {
	if controller.Env == nil {
		return os.Getenv(key)
	}
	return controller.Env(key)
}

func (controller *Controller) sanitizedEnvironment() []string {
	values := make(map[string]string)
	for _, entry := range os.Environ() {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		if strings.HasPrefix(key, "GUARDIAN_KEY_") || strings.HasPrefix(key, "ACTIONS_") ||
			strings.HasPrefix(key, "INPUT_") || key == "GH_TOKEN" ||
			key == "GITHUB_TOKEN" || key == "GUARDIAN_PROVIDERS" {
			continue
		}
		values[key] = value
	}
	values["GIT_TERMINAL_PROMPT"] = "0"
	result := make([]string, 0, len(values))
	for key, value := range values {
		result = append(result, key+"="+value)
	}
	sort.Strings(result)
	return result
}

type CommandResult struct {
	Output   string
	ExitCode int
}

type limitedBuffer struct {
	buffer bytes.Buffer
	limit  int
}

func (writer *limitedBuffer) Write(data []byte) (int, error) {
	original := len(data)
	remaining := writer.limit - writer.buffer.Len()
	if remaining > 0 {
		if len(data) > remaining {
			data = data[:remaining]
		}
		_, _ = writer.buffer.Write(data)
	}
	return original, nil
}

func (controller *Controller) run(args []string, cwd string, timeout time.Duration, check bool) (CommandResult, error) {
	if len(args) == 0 || timeout <= 0 {
		return CommandResult{}, fmt.Errorf("guardian: invalid command")
	}
	if cwd == "" {
		cwd = controller.Root
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	command := exec.CommandContext(ctx, args[0], args[1:]...)
	command.Dir = cwd
	command.Env = controller.sanitizedEnvironment()
	output := &limitedBuffer{limit: 8_000_000}
	command.Stdout = output
	command.Stderr = output
	err := command.Run()
	result := CommandResult{Output: output.buffer.String(), ExitCode: 0}
	if err != nil {
		result.ExitCode = 1
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			result.ExitCode = exit.ExitCode()
		}
		if ctx.Err() != nil {
			result.ExitCode = 124
			err = fmt.Errorf("guardian: command timed out")
		}
		if check {
			joined := strings.Join(args[:min(len(args), 4)], " ")
			message := result.Output
			if len(message) > 8000 {
				message = message[len(message)-8000:]
			}
			return result, fmt.Errorf("guardian: command failed: %s\n%s", joined, message)
		}
	}
	return result, nil
}

func (controller *Controller) runOK(args []string, cwd string, timeout time.Duration) (string, error) {
	result, err := controller.run(args, cwd, timeout, true)
	return result.Output, err
}

func (controller *Controller) runUnchecked(args []string, cwd string, timeout time.Duration) CommandResult {
	result, _ := controller.run(args, cwd, timeout, false)
	return result
}

func readJSONFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func writeJSONFile(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return writeFileAtomic(path, data, 0o644)
}

func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".guardian-write-*")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if err := temp.Chmod(mode); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempName, path)
}

func (controller *Controller) sourceFile(source string) (string, error) {
	target := filepath.Join(controller.Out, "bot-api.html")
	if source != "" {
		data, err := os.ReadFile(source)
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return "", err
		}
		return target, nil
	}
	client := &http.Client{Timeout: 45 * time.Second}
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		request, err := http.NewRequest(http.MethodGet, "https://core.telegram.org/bots/api", nil)
		if err != nil {
			return "", err
		}
		request.Header.Set("User-Agent", "hermes-guardian")
		response, err := client.Do(request)
		if err == nil {
			data, readErr := io.ReadAll(io.LimitReader(response.Body, 4_000_001))
			response.Body.Close()
			if readErr != nil {
				err = readErr
			} else if response.StatusCode != http.StatusOK {
				err = fmt.Errorf("official document returned HTTP %d", response.StatusCode)
			} else if len(data) > 4_000_000 {
				err = fmt.Errorf("official document exceeds size limit")
			} else {
				if writeErr := os.WriteFile(target, data, 0o644); writeErr != nil {
					return "", writeErr
				}
				return target, nil
			}
		}
		lastErr = err
		if attempt < 2 {
			time.Sleep(time.Duration(1<<attempt) * time.Second)
		}
	}
	_ = lastErr
	return "", fmt.Errorf("guardian: cannot download official API; previous state is preserved")
}

func (controller *Controller) buildTools() error {
	if err := os.MkdirAll(controller.Out, 0o755); err != nil {
		return err
	}
	for _, name := range []string{"botapi-schema", "botapi-diff", "botapi-models", "botapi-audit", "api-surface"} {
		_, err := controller.runOK([]string{"go", "build", "-o", filepath.Join(controller.Out, name), "./internal/cmd/" + name}, controller.Root, 180*time.Second)
		if err != nil {
			return err
		}
	}
	return nil
}

func (controller *Controller) tool(name string, cwd string, check bool, args ...string) (CommandResult, error) {
	command := []string{filepath.Join(controller.Out, name)}
	command = append(command, args...)
	return controller.run(command, cwd, 180*time.Second, check)
}

func isGeneratedPath(name string) bool {
	_, ok := generatedPaths[name]
	return ok
}

func safeRepositoryFile(root, name string) (string, error) {
	if name == "" || strings.Contains(name, "\\\\") || path.IsAbs(name) ||
		path.Clean(name) != name || name == "." || name == ".." || strings.HasPrefix(name, "../") {
		return "", fmt.Errorf("guardian: candidate path is invalid")
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	rootReal, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return "", err
	}
	current := rootReal
	parts := strings.Split(name, "/")
	for index, part := range parts {
		if part == "" || part == "." || part == ".." {
			return "", fmt.Errorf("guardian: candidate path is invalid")
		}
		current = filepath.Join(current, filepath.FromSlash(part))
		info, err := os.Lstat(current)
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("guardian: candidate path contains a symlink: %s", name)
		}
		if index < len(parts)-1 && !info.IsDir() {
			return "", fmt.Errorf("guardian: candidate parent is not a directory: %s", name)
		}
		if index == len(parts)-1 && !info.Mode().IsRegular() {
			return "", fmt.Errorf("guardian: candidate is not a regular file: %s", name)
		}
	}
	rel, err := filepath.Rel(rootReal, current)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("guardian: candidate path escapes repository")
	}
	return current, nil
}

func (controller *Controller) scopeCheck(root, base string) ([]string, error) {
	changed, err := controller.runOK([]string{"git", "diff", "--name-only", base}, root, 120*time.Second)
	if err != nil {
		return nil, err
	}
	untracked, err := controller.runOK([]string{"git", "ls-files", "--others", "--exclude-standard"}, root, 120*time.Second)
	if err != nil {
		return nil, err
	}
	names := make(map[string]struct{})
	for _, value := range append(strings.Split(changed, "\n"), strings.Split(untracked, "\n")...) {
		if value != "" {
			names[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(names))
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	for name := range names {
		existingResult := controller.runUnchecked([]string{"git", "cat-file", "-e", base + ":" + name}, root, 30*time.Second)
		existed := existingResult.ExitCode == 0
		if !isGeneratedPath(name) && !AllowedRepairPath(name, existed) {
			return nil, fmt.Errorf("guardian: candidate modified protected file: %s", name)
		}
		filePath, pathErr := safeRepositoryFile(rootAbs, name)
		if pathErr != nil {
			return nil, fmt.Errorf("guardian: candidate deleted or replaced a file: %s", name)
		}
		info, statErr := os.Stat(filePath)
		if statErr != nil {
			return nil, fmt.Errorf("guardian: candidate deleted or replaced a file: %s", name)
		}
		if info.Size() > 2_000_000 {
			return nil, fmt.Errorf("guardian: candidate file too large: %s", name)
		}
		result = append(result, name)
	}
	sort.Strings(result)
	return result, nil
}

func (controller *Controller) candidateDigest(root, base string) (string, error) {
	names, err := controller.scopeCheck(root, base)
	if err != nil {
		return "", err
	}
	content := make(map[string]string, len(names))
	for _, name := range names {
		if name == ".github/guardian/update.json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return "", err
		}
		content[name] = string(data)
	}
	return Digest(content)
}

func (controller *Controller) validationImage(minimum bool) (string, error) {
	version := ""
	if minimum {
		version = "1.25"
	} else {
		value, err := controller.runOK([]string{"go", "env", "GOVERSION"}, controller.Root, 30*time.Second)
		if err != nil {
			return "", err
		}
		version = strings.TrimPrefix(strings.TrimSpace(value), "go")
	}
	if !goVersionPattern.MatchString(version) {
		return "", fmt.Errorf("guardian: validation requires a stable Go toolchain")
	}
	image := "hermes-guardian:" + version
	if !minimum && controller.getenv("GUARDIAN_IMAGE") != "" {
		image = controller.getenv("GUARDIAN_IMAGE")
	}
	if !imagePattern.MatchString(image) {
		return "", fmt.Errorf("guardian: invalid validation image")
	}
	inspect := controller.runUnchecked([]string{"docker", "image", "inspect", image}, controller.Root, 60*time.Second)
	if inspect.ExitCode != 0 {
		if _, err := controller.runOK([]string{"docker", "build", "--build-arg", "GO_VERSION=" + strings.TrimPrefix(image, "hermes-guardian:"),
			"--tag", image, filepath.Join(controller.Root, "scripts", "guardian")}, controller.Root, 600*time.Second); err != nil {
			return "", err
		}
	}
	return image, nil
}

type SandboxResult struct {
	Passed bool   `json:"passed"`
	Output string `json:"output"`
}

func (controller *Controller) sandbox(command string, network bool, timeout time.Duration, minimum bool) (SandboxResult, error) {
	image, err := controller.validationImage(minimum)
	if err != nil {
		return SandboxResult{}, err
	}
	name := fmt.Sprintf("hermes-guardian-%d-%d", os.Getpid(), time.Now().UnixNano())
	networkMode := "none"
	if network {
		networkMode = "bridge"
	}
	args := []string{"run", "--rm", "--name", name, "--read-only", "--cap-drop=ALL",
		"--security-opt=no-new-privileges", "--pids-limit=512", "--cpus=2", "--memory=4g",
		"--network=" + networkMode, "--tmpfs", "/tmp:rw,exec,size=3g", "--user", "65534:65534",
		"-v", controller.Work + ":/repo:ro", "-w", "/repo",
		"-e", "HOME=/tmp", "-e", "GOCACHE=/tmp/go-cache", "-e", "GOPATH=/tmp/go",
		"-e", "GOTOOLCHAIN=local", "-e", "GOFLAGS=-buildvcs=false", "-e", "RELEASE_ALLOW_DIRTY=1",
		"-e", "GOMAXPROCS=2", image, "bash", "-c",
		"git config --global --add safe.directory /repo && " + command}
	logName := "repair-check.log"
	if minimum {
		logName = "minimum-go.log"
	} else if network {
		logName = "release.log"
	}
	logPath := filepath.Join(controller.Out, logName)
	if err := os.MkdirAll(controller.Out, 0o755); err != nil {
		return SandboxResult{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = controller.Root
	cmd.Env = controller.sanitizedEnvironment()
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return SandboxResult{}, err
	}
	writer := &limitedWriter{writer: file, remaining: 16_000_000}
	cmd.Stdout = writer
	cmd.Stderr = writer
	runErr := cmd.Run()
	closeErr := file.Close()
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cleanupCancel()
	cleanup := exec.CommandContext(cleanupCtx, "docker", "rm", "-f", name)
	cleanup.Env = controller.sanitizedEnvironment()
	_ = cleanup.Run()
	if closeErr != nil {
		return SandboxResult{}, closeErr
	}
	data, readErr := os.ReadFile(logPath)
	if readErr != nil {
		return SandboxResult{}, readErr
	}
	if ctx.Err() != nil {
		return SandboxResult{Passed: false, Output: "Validation exceeded its time limit"}, nil
	}
	if runErr != nil {
		return SandboxResult{Passed: false, Output: tailText(string(data), 16_000)}, nil
	}
	return SandboxResult{Passed: true, Output: tailText(string(data), 16_000)}, nil
}

type limitedWriter struct {
	writer    io.Writer
	remaining int
}

func (writer *limitedWriter) Write(data []byte) (int, error) {
	original := len(data)
	if writer.remaining <= 0 {
		return original, nil
	}
	if len(data) > writer.remaining {
		data = data[:writer.remaining]
	}
	n, err := writer.writer.Write(data)
	writer.remaining -= n
	if err != nil {
		return n, err
	}
	return original, nil
}

func tailText(value string, limit int) string {
	if len(value) > limit {
		return value[len(value)-limit:]
	}
	return value
}

func (controller *Controller) regenerate() (bool, error) {
	if _, err := controller.tool("botapi-models", controller.Work, true, "-root", controller.Work); err != nil {
		return false, err
	}
	if _, err := controller.tool("api-surface", controller.Work, true, "-output", "spec/api-surface.txt"); err != nil {
		return false, err
	}
	audit, err := controller.tool("botapi-audit", controller.Work, true, "-root", controller.Work,
		"-spec", filepath.Join(controller.Work, "spec/bot-api.json"), "-json", "-allow-gaps")
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(filepath.Join(controller.Out, "audit.json"), []byte(audit.Output), 0o644); err != nil {
		return false, err
	}
	result, err := controller.tool("botapi-audit", controller.Work, false, "-root", controller.Work,
		"-spec", filepath.Join(controller.Work, "spec/bot-api.json"))
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(filepath.Join(controller.Out, "audit.txt"), []byte(result.Output), 0o644); err != nil {
		return false, err
	}
	return result.ExitCode == 0, nil
}

func (controller *Controller) summary(report map[string]any) string {
	status := strings.ReplaceAll(stringValue(report["status"]), "_", " ")
	lines := []string{"## Hermes maintenance", "", "**" + status + "**", ""}
	for _, key := range []string{"version", "release", "branch", "pr", "message"} {
		if value := report[key]; value != nil && fmt.Sprint(value) != "" {
			lines = append(lines, "- "+strings.ToUpper(key[:1])+key[1:]+": "+fmt.Sprint(value))
		}
	}
	if reasons := arrayValue(report["semantic_reasons"]); len(reasons) > 0 {
		lines = append(lines, "", "Documentation changes:")
		for _, reason := range reasons[:min(30, len(reasons))] {
			lines = append(lines, "- "+fmt.Sprint(reason))
		}
	}
	if repair := mapValue(report["repair"]); len(repair) > 0 {
		lines = append(lines, "", "Repair: "+stringValue(repair["summary"]))
	}
	if review := mapValue(report["review"]); len(review) > 0 {
		lines = append(lines, "", "Independent review: "+stringValue(review["summary"]))
	}
	lines = append(lines, "", "Reports include the official delta and exact validation result. Existing applications keep their installed Hermes version.")
	return strings.Join(lines, "\n") + "\n"
}

func (controller *Controller) finish(report map[string]any) error {
	if err := writeJSONFile(filepath.Join(controller.Out, "report.json"), report); err != nil {
		return err
	}
	summary := controller.summary(report)
	if err := os.WriteFile(filepath.Join(controller.Out, "summary.md"), []byte(summary), 0o644); err != nil {
		return err
	}
	if target := controller.getenv("GITHUB_STEP_SUMMARY"); target != "" {
		file, err := os.OpenFile(target, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		_, writeErr := file.WriteString(summary)
		closeErr := file.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	fmt.Println(stringValue(report["status"]) + ": " + stringValue(report["message"]))
	return nil
}
