package guardian

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sidwiskers/hermes/internal/botapi"
)

const guardianStatePath = ".github/guardian/update.json"

func (controller *Controller) Prepare(source string, retry bool) error {
	if err := os.RemoveAll(controller.Out); err != nil {
		return err
	}
	if err := os.MkdirAll(controller.Out, 0o755); err != nil {
		return err
	}
	baseOutput, err := controller.runOK([]string{"git", "rev-parse", "HEAD"}, controller.Root, 30*time.Second)
	if err != nil {
		return err
	}
	base := strings.TrimSpace(baseOutput)
	report := map[string]any{
		"format":  1,
		"base":    base,
		"status":  "checking",
		"created": time.Now().UTC().Format(time.RFC3339Nano),
	}
	if controller.getenv("GUARDIAN_MODE") == "paused" {
		report["status"] = "paused"
		report["message"] = "Maintenance is paused in repository variables."
		return controller.finish(report)
	}
	if err := controller.buildTools(); err != nil {
		return err
	}
	htmlPath, err := controller.sourceFile(source)
	if err != nil {
		return err
	}
	candidatePath := filepath.Join(controller.Out, "bot-api.json")
	if _, err := controller.tool("botapi-schema", controller.Root, true, "-source", htmlPath, "-output", candidatePath); err != nil {
		return err
	}
	before, err := botapi.Load(filepath.Join(controller.Root, "spec", "bot-api.json"))
	if err != nil {
		return fmt.Errorf("guardian: load previous manifest: %w", err)
	}
	candidate, err := botapi.Load(candidatePath)
	if err != nil {
		return fmt.Errorf("guardian: load candidate manifest: %w", err)
	}
	if err := ValidateCandidate(before, candidate); err != nil {
		return err
	}
	diffOutput, err := controller.tool("botapi-diff", controller.Root, true, "-after", candidatePath, "-format", "json")
	if err != nil {
		return err
	}
	var diff botapi.SurfaceDiff
	if err := json.Unmarshal([]byte(diffOutput.Output), &diff); err != nil {
		return fmt.Errorf("guardian: decode schema diff: %w", err)
	}
	html, err := os.ReadFile(htmlPath)
	if err != nil {
		return err
	}
	semantics, err := SnapshotHTML(html)
	if err != nil {
		return err
	}
	previousSemanticsPath := filepath.Join(controller.Root, "spec", "bot-api-semantics.json")
	var previousSemantics *Snapshot
	if _, err := os.Stat(previousSemanticsPath); err == nil {
		data, readErr := os.ReadFile(previousSemanticsPath)
		if readErr != nil {
			return readErr
		}
		var previous Snapshot
		if err := json.Unmarshal(data, &previous); err != nil {
			return fmt.Errorf("guardian: decode previous semantic baseline: %w", err)
		}
		previousSemantics = &previous
	} else if !os.IsNotExist(err) {
		return err
	}
	reasons := SemanticChanges(previousSemantics, semantics, diff)
	if err := writeJSONFile(filepath.Join(controller.Out, "diff.json"), diff); err != nil {
		return err
	}
	markdown, err := controller.tool("botapi-diff", controller.Root, true, "-after", candidatePath, "-format", "markdown")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(controller.Out, "diff.md"), []byte(markdown.Output), 0o644); err != nil {
		return err
	}
	report["version"] = candidate.Version
	report["semantic_reasons"] = stringSliceAny(reasons)
	if !diff.Changed && len(reasons) == 0 {
		report["status"] = "unchanged"
		report["message"] = "Hermes matches Telegram's API and documented behavior."
		return controller.finish(report)
	}

	schemaIdentity := make(map[string]any)
	candidateJSON, err := os.ReadFile(candidatePath)
	if err != nil {
		return err
	}
	var candidateValue map[string]any
	if err := json.Unmarshal(candidateJSON, &candidateValue); err != nil {
		return err
	}
	for key, value := range candidateValue {
		if key != "source_sha256" && key != "released" {
			schemaIdentity[key] = value
		}
	}
	identity, err := Digest(map[string]any{"schema": schemaIdentity, "semantics": semantics})
	if err != nil {
		return err
	}
	branch := "automation/bot-api-" + candidate.Version + "-" + identity[:12]
	report["branch"] = branch
	report["identity"] = identity
	report["mechanical"] = diff.Classification == "mechanical" && len(reasons) == 0

	if err := os.RemoveAll(controller.Work); err != nil {
		return err
	}
	if _, err := controller.runOK([]string{"git", "clone", "--no-hardlinks", controller.Root, controller.Work}, controller.Root, 180*time.Second); err != nil {
		return err
	}
	for _, pair := range [][2]string{{"user.name", "hermes-guardian[bot]"}, {"user.email", "hermes-guardian[bot]@users.noreply.github.com"}, {"core.hooksPath", "/dev/null"}} {
		if _, err := controller.runOK([]string{"git", "config", pair[0], pair[1]}, controller.Work, 30*time.Second); err != nil {
			return err
		}
	}

	var expected any
	previousState := map[string]any{}
	repository := controller.getenv("GITHUB_REPOSITORY")
	if repository != "" {
		if !repositoryPattern.MatchString(repository) {
			return fmt.Errorf("guardian: invalid GITHUB_REPOSITORY")
		}
		remote := "https://github.com/" + repository + ".git"
		found := controller.runUnchecked([]string{"git", "ls-remote", "--exit-code", remote, "refs/heads/" + branch}, controller.Work, 60*time.Second)
		if found.ExitCode != 0 && found.ExitCode != 2 {
			return fmt.Errorf("guardian: cannot check existing update branch; refusing to overwrite work")
		}
		if found.ExitCode == 0 {
			fields := strings.Fields(found.Output)
			if len(fields) == 0 {
				return fmt.Errorf("guardian: invalid existing update branch response")
			}
			expected = fields[0]
			if _, err := controller.runOK([]string{"git", "fetch", remote, "refs/heads/" + branch}, controller.Work, 180*time.Second); err != nil {
				return err
			}
			if _, err := controller.runOK([]string{"git", "checkout", "--detach", "FETCH_HEAD"}, controller.Work, 60*time.Second); err != nil {
				return err
			}
			ancestorOutput, err := controller.runOK([]string{"git", "merge-base", base, "HEAD"}, controller.Work, 60*time.Second)
			if err != nil {
				return err
			}
			ancestor := strings.TrimSpace(ancestorOutput)
			existingPaths, err := controller.scopeCheck(controller.Work, ancestor)
			if err != nil {
				return err
			}
			for _, path := range existingPaths {
				name := filepath.Base(path)
				if strings.HasSuffix(path, ".go") && !strings.HasPrefix(name, "zz_") && path != "api/client.go" {
					report["mechanical"] = false
				}
			}
			originalClient, err := controller.runOK([]string{"git", "show", ancestor + ":api/client.go"}, controller.Work, 30*time.Second)
			if err != nil {
				return err
			}
			currentClient, err := os.ReadFile(filepath.Join(controller.Work, "api", "client.go"))
			if err != nil {
				return err
			}
			stripVersion := regexp.MustCompile(`hermes-go/[0-9]+\.[0-9]+\.[0-9]+`).ReplaceAllString
			if stripVersion(originalClient, "hermes-go/VERSION") != stripVersion(string(currentClient), "hermes-go/VERSION") {
				report["mechanical"] = false
			}
			merged := controller.runUnchecked([]string{"git", "merge", "--no-edit", base}, controller.Work, 120*time.Second)
			if merged.ExitCode != 0 {
				report["status"] = "needs_attention"
				report["message"] = "The existing update conflicts with main; its branch is preserved."
				return controller.finish(report)
			}
			if _, err := controller.scopeCheck(controller.Work, base); err != nil {
				return err
			}
			statePath := filepath.Join(controller.Work, filepath.FromSlash(guardianStatePath))
			if _, err := os.Stat(statePath); err == nil {
				previousState, err = readJSONFile(statePath)
				if err != nil {
					return err
				}
			} else if !os.IsNotExist(err) {
				return err
			}
		}
	}
	report["expected_head"] = expected
	if oldIdentity := stringValue(previousState["identity"]); oldIdentity != "" && oldIdentity != identity {
		return fmt.Errorf("guardian: update branch contains a different official source identity")
	}
	if stringValue(previousState["identity"]) == identity {
		oldManifest, err := botapi.Load(filepath.Join(controller.Work, "spec", "bot-api.json"))
		if err != nil {
			return err
		}
		candidate.SourceSHA256 = oldManifest.SourceSHA256
		encoded, err := botapi.Marshal(candidate)
		if err != nil {
			return err
		}
		if err := os.WriteFile(candidatePath, encoded, 0o644); err != nil {
			return err
		}
	}
	currentDigest, err := controller.candidateDigest(controller.Work, base)
	if err != nil {
		return err
	}
	sameCandidate := previousState["last_base"] == base && previousState["candidate_digest"] == currentDigest
	if sameCandidate && !retry && stringValue(previousState["status"]) == "reviewed" {
		report["status"] = "reviewed"
		report["repair"] = previousState["repair"]
		report["review"] = previousState["review"]
		report["message"] = "The previously verified repair awaits your PR review. Its tested code is unchanged."
		return controller.finish(report)
	}
	repairRuns := intValue(previousState["repair_runs"])
	if repairRuns < 0 || repairRuns > 100 {
		return fmt.Errorf("guardian: invalid persisted repair count")
	}
	if sameCandidate && !retry && stringValue(previousState["status"]) == "needs_attention" &&
		(repairRuns >= 2 || controller.getenv("GUARDIAN_REPAIR") != "true" || !controller.hasRepairProviders()) {
		report["status"] = "needs_attention"
		report["repair"] = previousState["repair"]
		report["review"] = previousState["review"]
		report["message"] = "The prepared update still needs attention. Its branch is preserved; use Retry repair to resume."
		return controller.finish(report)
	}

	if err := copyFile(candidatePath, filepath.Join(controller.Work, "spec", "bot-api.json")); err != nil {
		return err
	}
	if err := writeJSONFile(filepath.Join(controller.Work, "spec", "bot-api-semantics.json"), semantics); err != nil {
		return err
	}
	parity, generationErr := controller.regenerate()
	if generationErr != nil {
		message := generationErr.Error()
		if len(message) > 4000 {
			message = message[len(message)-4000:]
		}
		report["generation_error"] = message
	}
	report["parity"] = parity
	if !parity {
		report["mechanical"] = false
	}

	changedSections := make(map[string]any)
	reasonSections := make(map[string]struct{})
	for _, reason := range reasons {
		name, _, _ := strings.Cut(reason, ":")
		reasonSections[name] = struct{}{}
	}
	for _, category := range []botapi.NamedChanges{diff.Methods, diff.Objects, diff.Unions} {
		for _, name := range append(append([]string{}, category.Added...), category.Changed...) {
			reasonSections[strings.ToLower(name)] = struct{}{}
		}
	}
	for name, section := range semantics.Sections {
		if _, ok := reasonSections[name]; ok {
			changedSections[name] = section
		}
	}
	evidence := map[string]any{
		"diff": diff,
		"semantic_reasons": reasons,
		"release_announcements": semantics.Announcements,
		"official_sections": changedSections,
		"previous_review": previousState["review"],
		"previous_validation": previousState["validation"],
	}
	if data, err := os.ReadFile(filepath.Join(controller.Out, "audit.json")); err == nil {
		var audit any
		if json.Unmarshal(data, &audit) == nil {
			evidence["audit"] = audit
		}
	} else {
		evidence["audit"] = report["generation_error"]
	}

	if !boolValue(report["mechanical"]) || !parity {
		report["repair"] = previousState["repair"]
		if report["repair"] == nil {
			report["repair"] = map[string]any{"status": "disabled", "summary": "Coding repair is disabled."}
		}
		if controller.getenv("GUARDIAN_REPAIR") == "true" && (repairRuns < 2 || retry) {
			if _, err := exec.LookPath("docker"); err != nil {
				return fmt.Errorf("guardian: Docker is required before running coding repair")
			}
			if _, err := controller.validationImage(false); err != nil {
				return err
			}
			protectedTests, err := controller.gitTreePaths(base)
			if err != nil {
				return err
			}
			repairResult := Repair(controller.Work, evidence, func() any {
				result, err := controller.sandbox("go test -shuffle=on -count=1 ./...", false, 300*time.Second, false)
				if err != nil {
					return map[string]any{"passed": false, "output": err.Error()}
				}
				return result
			}, 24, protectedTests, controller.getenv)
			report["repair"] = repairResult
			if repairResult.Status != "unconfigured" {
				repairRuns++
			}
			paths, err := controller.scopeCheck(controller.Work, base)
			if err != nil {
				return err
			}
			goFiles := make([]string, 0)
			for _, name := range paths {
				if strings.HasSuffix(name, ".go") {
					goFiles = append(goFiles, filepath.Join(controller.Work, filepath.FromSlash(name)))
				}
			}
			if len(goFiles) > 0 {
				if _, err := controller.runOK(append([]string{"gofmt", "-w"}, goFiles...), controller.Work, 120*time.Second); err != nil {
					report["generation_error"] = tailText(err.Error(), 4000)
					parity = false
				} else {
					parity, generationErr = controller.regenerate()
					if generationErr != nil {
						report["generation_error"] = tailText(generationErr.Error(), 4000)
					}
				}
			}
			report["parity"] = parity
			if _, err := controller.runOK([]string{"git", "add", "--all"}, controller.Work, 60*time.Second); err != nil {
				return err
			}
			patch, err := controller.runOK([]string{"git", "diff", "--cached", base, "--", "*.go", "docs/"}, controller.Work, 60*time.Second)
			if err != nil {
				return err
			}
			finalTests, testErr := controller.sandbox("go test -shuffle=on -count=1 ./...", false, 300*time.Second, false)
			if testErr != nil {
				evidence["final_tests"] = map[string]any{"passed": false, "output": testErr.Error()}
			} else {
				evidence["final_tests"] = finalTests
			}
			if len(patch.Output) <= 250_000 {
				report["review"] = Review(controller.Work, evidence, patch.Output, controller.getenv)
			} else {
				report["review"] = map[string]any{"approved": false, "summary": "Patch exceeds automated review budget"}
			}
		} else if repairRuns >= 2 {
			report["message"] = "Repair reached its two-run budget. Use Retry repair after addressing the reported issue."
		}
	}
	report["repair_runs"] = repairRuns
	report["status"] = "prepared"
	if _, ok := report["message"]; !ok {
		report["message"] = "Update prepared; validation decides readiness."
	}
	if report["review"] == nil {
		if oldReview := mapValue(previousState["review"]); len(oldReview) > 0 {
			copied := make(map[string]any, len(oldReview))
			for key, value := range oldReview {
				copied[key] = value
			}
			copied["approved"] = false
			report["review"] = copied
		}
	}
	state := map[string]any{}
	for _, key := range []string{"identity", "repair_runs", "repair", "review"} {
		if value, ok := report[key]; ok {
			state[key] = value
		}
	}
	if err := writeJSONFile(filepath.Join(controller.Work, filepath.FromSlash(guardianStatePath)), state); err != nil {
		return err
	}
	if _, err := controller.scopeCheck(controller.Work, base); err != nil {
		return err
	}
	return controller.finish(report)
}

func (controller *Controller) hasRepairProviders() bool {
	providers, err := ProvidersFromEnvironment(controller.getenv)
	return err == nil && len(providers) > 0
}

func (controller *Controller) gitTreePaths(commit string) (map[string]struct{}, error) {
	output, err := controller.runOK([]string{"git", "ls-tree", "-r", "--name-only", commit}, controller.Root, 60*time.Second)
	if err != nil {
		return nil, err
	}
	paths := make(map[string]struct{})
	for _, name := range strings.Split(strings.TrimSpace(output), "\n") {
		if name != "" {
			paths[name] = struct{}{}
		}
	}
	return paths, nil
}

func (controller *Controller) Validate() error {
	report, err := readJSONFile(filepath.Join(controller.Out, "report.json"))
	if err != nil {
		return err
	}
	if stringValue(report["status"]) != "prepared" {
		return nil
	}
	base := stringValue(report["base"])
	parity := boolValue(report["parity"])
	if parity {
		if err := controller.metadata(report); err != nil {
			return controller.validationAttention(report, err)
		}
		if _, err := controller.regenerate(); err != nil {
			return controller.validationAttention(report, err)
		}
	}
	if _, err := controller.scopeCheck(controller.Work, base); err != nil {
		return controller.validationAttention(report, err)
	}
	result, err := controller.sandbox("./scripts/release-check.sh", true, 2100*time.Second, false)
	if err != nil {
		return controller.validationAttention(report, err)
	}
	if result.Passed {
		minimum, minErr := controller.sandbox("./scripts/check-generated.sh && go vet ./... && go test -shuffle=on -count=1 ./...",
			false, 600*time.Second, true)
		if minErr != nil {
			return controller.validationAttention(report, minErr)
		}
		result = SandboxResult{Passed: minimum.Passed, Output: result.Output + "\nGo 1.25 verification:\n" + tailText(minimum.Output, 4000)}
	}
	report["validation"] = result
	status := "needs_attention"
	if result.Passed && parity {
		status = "ready"
	}
	if status == "ready" && !boolValue(report["mechanical"]) {
		if boolValue(mapValue(report["review"])["approved"]) {
			status = "reviewed"
		} else {
			status = "needs_attention"
		}
	}
	report["status"] = status
	switch status {
	case "ready":
		report["message"] = "Generated update passed the complete release gates."
	case "reviewed":
		report["message"] = "Repaired update passed release gates and a separate AI review; maintainer release approval is required."
	default:
		report["message"] = "Update is preserved as a draft. See validation and repair evidence for the remaining work."
	}
	state := map[string]any{}
	for _, key := range []string{"identity", "repair_runs", "repair", "review", "status"} {
		if value, ok := report[key]; ok {
			state[key] = value
		}
	}
	state["last_base"] = base
	candidateDigest, err := controller.candidateDigest(controller.Work, base)
	if err != nil {
		return err
	}
	state["candidate_digest"] = candidateDigest
	if status == "needs_attention" {
		state["validation"] = report["validation"]
	}
	if err := writeJSONFile(filepath.Join(controller.Work, filepath.FromSlash(guardianStatePath)), state); err != nil {
		return err
	}
	paths, err := controller.scopeCheck(controller.Work, base)
	if err != nil {
		return err
	}
	changes := make([]map[string]any, 0, len(paths))
	totalBytes := 0
	for _, name := range paths {
		data, err := os.ReadFile(filepath.Join(controller.Work, filepath.FromSlash(name)))
		if err != nil {
			return err
		}
		totalBytes += len(data)
		changes = append(changes, map[string]any{"path": name, "content": string(data), "mode": "100644"})
	}
	bundle := map[string]any{"report": report, "changes": changes}
	encoded, err := CanonicalJSON(bundle)
	if err != nil {
		return err
	}
	if len(encoded) > 6_000_000 || totalBytes > 6_000_000 {
		return fmt.Errorf("guardian: update bundle exceeds publication limit")
	}
	if err := writeJSONFile(filepath.Join(controller.Out, "bundle.json"), bundle); err != nil {
		return err
	}
	return controller.finish(report)
}

func (controller *Controller) validationAttention(report map[string]any, cause error) error {
	message := cause.Error()
	report["status"] = "needs_attention"
	report["message"] = tailText(message, 6000)
	if err := controller.finish(report); err != nil {
		return err
	}
	return nil
}

func (controller *Controller) metadata(report map[string]any) error {
	tags, err := controller.runOK([]string{"git", "tag", "--list", "v*"}, controller.Root, 60*time.Second)
	if err != nil {
		return err
	}
	var latest version
	found := false
	versionPattern := regexp.MustCompile(`^v([0-9]+)\.([0-9]+)\.([0-9]+)$`)
	for _, tag := range strings.Split(tags, "\n") {
		match := versionPattern.FindStringSubmatch(tag)
		if len(match) != 4 {
			continue
		}
		major, _ := strconv.Atoi(match[1])
		minor, _ := strconv.Atoi(match[2])
		candidate := version{major: major, minor: minor}
		if !found || compareVersion(candidate, latest) > 0 {
			latest = candidate
			found = true
		}
	}
	if !found {
		return fmt.Errorf("guardian: no stable release tag found")
	}
	newVersion := fmt.Sprintf("%d.%d.0", latest.major, latest.minor+1)
	report["release"] = "v" + newVersion

	changelogPath := filepath.Join(controller.Work, "CHANGELOG.md")
	changelog, err := os.ReadFile(changelogPath)
	if err != nil {
		return err
	}
	body := string(changelog)
	if !strings.Contains(body, "## "+newVersion+" -") {
		date := time.Now().UTC().Format("2006-01-02")
		entry := fmt.Sprintf("## %s - %s\n\n- Update the typed Telegram API and documentation baseline to Bot API %s.\n- See the Guardian update PR for the official delta, compatibility review and validation evidence.\n\n",
			newVersion, date, stringValue(report["version"]))
		body, _ = strings.Cut(strings.Replace(body, "# Changelog\n", "# Changelog\n\n"+entry, 1), "\x00")
		if !strings.Contains(body, entry) {
			return fmt.Errorf("guardian: unrecognized changelog format")
		}
		if err := os.WriteFile(changelogPath, []byte(body), 0o644); err != nil {
			return err
		}
	}
	clientPath := filepath.Join(controller.Work, "api", "client.go")
	clientBytes, err := os.ReadFile(clientPath)
	if err != nil {
		return err
	}
	userAgent := regexp.MustCompile(`hermes-go/[0-9]+\.[0-9]+\.[0-9]+`)
	updated := userAgent.ReplaceAll(clientBytes, []byte("hermes-go/"+newVersion))
	if bytesCount := len(userAgent.FindAll(clientBytes, -1)); bytesCount != 1 {
		return fmt.Errorf("guardian: unrecognized user agent version")
	}
	if err := os.WriteFile(clientPath, updated, 0o644); err != nil {
		return err
	}

	schema, err := botapi.Load(filepath.Join(controller.Work, "spec", "bot-api.json"))
	if err != nil {
		return err
	}
	stats := schema.Stats()
	docPath := filepath.Join(controller.Work, "docs", "schema-parity.md")
	docBytes, err := os.ReadFile(docPath)
	if err != nil {
		return err
	}
	doc := regexp.MustCompile(`API [0-9]+\.[0-9]+ and contains:`).ReplaceAllString(string(docBytes),
		"API "+stringValue(report["version"])+" and contains:")
	counts := []struct {
		label string
		value int
	}{
		{"Methods", stats.Methods},
		{"Method parameters", stats.Parameters},
		{"Concrete objects", stats.Objects},
		{"Object fields", stats.ObjectFields},
		{"Union roots", stats.Unions},
		{"Union variants", stats.Variants},
	}
	for _, count := range counts {
		pattern := regexp.MustCompile(`\| ` + regexp.QuoteMeta(count.label) + ` \| [0-9,]+ \| 0 \|`)
		replacement := fmt.Sprintf("| %s | %s | 0 |", count.label, commaNumber(count.value))
		doc = pattern.ReplaceAllString(doc, replacement)
	}
	return os.WriteFile(docPath, []byte(doc), 0o644)
}

func copyFile(source, target string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, data, 0o644)
}

func stringSliceAny(values []string) []any {
	result := make([]any, len(values))
	for index, value := range values {
		result[index] = value
	}
	return result
}

func commaNumber(value int) string {
	raw := strconv.Itoa(value)
	if len(raw) <= 3 {
		return raw
	}
	first := len(raw) % 3
	if first == 0 {
		first = 3
	}
	var builder strings.Builder
	builder.WriteString(raw[:first])
	for index := first; index < len(raw); index += 3 {
		builder.WriteByte(',')
		builder.WriteString(raw[index : index+3])
	}
	return builder.String()
}
