package guardian

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestAutomaticCandidateRequiresAllGatesAndGeneratedFiles(t *testing.T) {
	report := map[string]any{
		"status":     "ready",
		"mechanical": true,
		"parity":     true,
		"validation": map[string]any{"passed": true},
	}
	generated := []publicationChange{{Path: "types/zz_botapi_generated.go"}}
	if !AutomaticCandidate(report, generated) {
		t.Fatal("pure generated update did not qualify")
	}
	cases := []map[string]any{
		{"repair": map[string]any{"status": "proposed"}},
		{"parity": false},
		{"validation": map[string]any{"passed": false}},
		{"mechanical": false},
		{"status": "reviewed"},
	}
	for _, update := range cases {
		candidate := make(map[string]any, len(report)+len(update))
		for key, value := range report {
			candidate[key] = value
		}
		for key, value := range update {
			candidate[key] = value
		}
		if AutomaticCandidate(candidate, generated) {
			t.Errorf("unsafe candidate qualified with override %v", update)
		}
	}
	if AutomaticCandidate(report, []publicationChange{{Path: "api/handwritten.go"}}) {
		t.Fatal("handwritten source qualified for automatic merge")
	}
}

func TestHealthPausedNeverRequiresOrCallsGitHub(t *testing.T) {
	root := t.TempDir()
	controller := &Controller{
		Root: root,
		Out:  filepath.Join(root, ".guardian"),
		Env: func(key string) string {
			if key == "GUARDIAN_MODE" {
				return "paused"
			}
			return ""
		},
	}
	if err := controller.Health(); err != nil {
		t.Fatalf("paused health failed without GitHub credentials: %v", err)
	}
	report, err := readJSONFile(filepath.Join(controller.Out, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	if stringValue(report["status"]) != "paused" {
		t.Fatalf("paused health status = %q", stringValue(report["status"]))
	}
}

func TestPreparePausedRemovesStalePublicationBundle(t *testing.T) {
	root := t.TempDir()
	initGuardianGit(t, root, map[string]string{"README.md": "baseline"})
	controller := &Controller{
		Root: root,
		Out:  filepath.Join(root, ".guardian"),
		Work: filepath.Join(root, ".guardian", "work"),
		Env: func(key string) string {
			if key == "GUARDIAN_MODE" {
				return "paused"
			}
			return ""
		},
	}
	if err := os.MkdirAll(controller.Out, 0o755); err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(controller.Out, "bundle.json")
	if err := os.WriteFile(stale, []byte(`{"stale":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := controller.Prepare("", false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale publication bundle remains: %v", err)
	}
	report, err := readJSONFile(filepath.Join(controller.Out, "report.json"))
	if err != nil || stringValue(report["status"]) != "paused" {
		t.Fatalf("paused report = %#v, err=%v", report, err)
	}
}

func TestSubprocessEnvironmentExcludesCredentials(t *testing.T) {
	for key, value := range map[string]string{
		"GH_TOKEN":              "private",
		"GITHUB_TOKEN":          "private",
		"GUARDIAN_KEY_1":        "private",
		"ACTIONS_RUNTIME_TOKEN": "private",
		"INPUT_TOKEN":           "private",
		"GUARDIAN_PROVIDERS":    "private",
	} {
		t.Setenv(key, value)
	}
	controller := &Controller{Root: t.TempDir()}
	values := make(map[string]string)
	for _, entry := range controller.sanitizedEnvironment() {
		key, value, ok := strings.Cut(entry, "=")
		if ok {
			values[key] = value
		}
	}
	for _, key := range []string{"GH_TOKEN", "GITHUB_TOKEN", "GUARDIAN_KEY_1", "ACTIONS_RUNTIME_TOKEN", "INPUT_TOKEN", "GUARDIAN_PROVIDERS"} {
		if _, exists := values[key]; exists {
			t.Errorf("subprocess environment retained secret %q", key)
		}
	}
}

func TestGitTreePathsProtectsOnlyExistingTests(t *testing.T) {
	root := t.TempDir()
	base := initGuardianGit(t, root, map[string]string{
		"api/method.go":      "package api",
		"api/method_test.go": "package api",
		"README.md":          "docs",
	})
	controller := &Controller{Root: root}
	protected, err := controller.gitTreePaths(base)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := protected["api/method_test.go"]; !ok {
		t.Fatal("existing test file was not protected")
	}
	if _, ok := protected["api/method.go"]; ok {
		t.Fatal("ordinary source file was incorrectly protected")
	}
	if _, ok := protected["README.md"]; ok {
		t.Fatal("documentation was incorrectly protected")
	}
}

func TestScopeCheckPreservesExistingTestsButAllowsNewRegressionTests(t *testing.T) {
	root := t.TempDir()
	base := initGuardianGit(t, root, map[string]string{
		"api/method.go":        "package api",
		"api/existing_test.go": "package api",
	})
	controller := &Controller{Root: root}
	existingTest := filepath.Join(root, "api", "existing_test.go")
	if err := os.WriteFile(existingTest, []byte("package api\n// weakened"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := controller.scopeCheck(root, base); err == nil {
		t.Fatal("scope check accepted modification of an existing regression test")
	}
	if err := os.WriteFile(existingTest, []byte("package api"), 0o644); err != nil {
		t.Fatal(err)
	}
	newTest := filepath.Join(root, "api", "new_test.go")
	if err := os.WriteFile(newTest, []byte("package api"), 0o644); err != nil {
		t.Fatal(err)
	}
	paths, err := controller.scopeCheck(root, base)
	if err != nil {
		t.Fatalf("new regression test rejected: %v", err)
	}
	if !reflect.DeepEqual(paths, []string{"api/new_test.go"}) {
		t.Fatalf("new test paths = %v", paths)
	}
}

func TestPublishReleaseNeverMovesExistingTag(t *testing.T) {
	const existingSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/repos/owner/repo/git/ref/tags/v1.4.0" {
			t.Errorf("unexpected request path %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": map[string]any{"sha": existingSHA, "type": "commit"},
		})
	}))
	defer server.Close()
	controller := controllerWithGitHub(t, server.URL)
	err := controller.PublishRelease("v1.4.0", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "notes")
	if err == nil || !strings.Contains(err.Error(), "never move") {
		t.Fatalf("moving an existing release tag returned %v", err)
	}
	if calls != 1 {
		t.Fatalf("immutable tag check made %d requests, want 1", calls)
	}
}

func TestPublishReleaseIsIdempotent(t *testing.T) {
	const releaseSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		switch r.URL.Path {
		case "/repos/owner/repo/git/ref/tags/v1.4.0":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": map[string]any{"sha": releaseSHA, "type": "commit"},
			})
		case "/repos/owner/repo/releases/tags/v1.4.0":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 7, "draft": false, "prerelease": false})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	controller := controllerWithGitHub(t, server.URL)
	if err := controller.PublishRelease("v1.4.0", releaseSHA, "notes"); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"GET /repos/owner/repo/git/ref/tags/v1.4.0",
		"GET /repos/owner/repo/releases/tags/v1.4.0",
	}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("idempotent release requests = %v, want %v", paths, want)
	}
}

func TestPublishReleaseRecoversExistingTagWithoutMovingIt(t *testing.T) {
	const releaseSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		switch {
		case r.Method == "GET" && r.URL.Path == "/repos/owner/repo/git/ref/tags/v1.4.0":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": map[string]any{"sha": releaseSHA, "type": "commit"},
			})
		case r.Method == "GET" && r.URL.Path == "/repos/owner/repo/releases/tags/v1.4.0":
			http.NotFound(w, r)
		case r.Method == "POST" && r.URL.Path == "/repos/owner/repo/releases":
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Error(err)
			}
			if payload["tag_name"] != "v1.4.0" || payload["target_commitish"] != releaseSHA {
				t.Errorf("release payload = %#v", payload)
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 7})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	controller := controllerWithGitHub(t, server.URL)
	if err := controller.PublishRelease("v1.4.0", releaseSHA, "notes"); err != nil {
		t.Fatal(err)
	}
	if len(paths) != 3 || paths[2] != "POST /repos/owner/repo/releases" {
		t.Fatalf("partial-release recovery requests = %v", paths)
	}
}

func TestPublishRejectsConcurrentMainOrRepairChanges(t *testing.T) {
	const baseSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const changedSHA = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	root := t.TempDir()
	controller := &Controller{Root: root, Out: root}
	bundle := map[string]any{
		"report": map[string]any{
			"status":        "ready",
			"branch":        "automation/bot-api-10.4-aaaaaaaaaaaa",
			"base":          baseSHA,
			"expected_head": changedSHA,
		},
		"changes": []any{},
	}
	if err := writeJSONFile(filepath.Join(root, "bundle.json"), bundle); err != nil {
		t.Fatal(err)
	}

	t.Run("main changed", func(t *testing.T) {
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": map[string]any{"sha": changedSHA},
			})
		}))
		defer server.Close()
		controller.API = githubForServer(t, server.URL)
		err := controller.Publish()
		if err == nil || !strings.Contains(err.Error(), "main changed") {
			t.Fatalf("concurrent main update returned %v", err)
		}
		if calls != 1 {
			t.Fatalf("main race made %d requests, want 1", calls)
		}
	})

	t.Run("repair branch changed", func(t *testing.T) {
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			if strings.HasSuffix(r.URL.Path, "/git/ref/heads/main") {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"object": map[string]any{"sha": baseSHA},
				})
				return
			}
			if strings.HasSuffix(r.URL.Path, "/git/ref/heads/automation/bot-api-10.4-aaaaaaaaaaaa") {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"object": map[string]any{"sha": "cccccccccccccccccccccccccccccccccccccccc"},
				})
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()
		controller.API = githubForServer(t, server.URL)
		err := controller.Publish()
		if err == nil || !strings.Contains(err.Error(), "branch changed") {
			t.Fatalf("concurrent repair update returned %v", err)
		}
		if calls != 2 {
			t.Fatalf("branch race made %d requests, want 2", calls)
		}
	})
}

func controllerWithGitHub(t *testing.T, baseURL string) *Controller {
	t.Helper()
	return &Controller{API: githubForServer(t, baseURL)}
}

func githubForServer(t *testing.T, baseURL string) *GitHubClient {
	t.Helper()
	client, err := NewGitHubClient("owner/repo", "test-token")
	if err != nil {
		t.Fatal(err)
	}
	client.BaseURL = baseURL
	return client
}

func initGuardianGit(t *testing.T, root string, files map[string]string) string {
	t.Helper()
	run := func(args ...string) string {
		command := exec.Command("git", args...)
		command.Dir = root
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, output)
		}
		return strings.TrimSpace(string(output))
	}
	run("init")
	run("config", "user.name", "Guardian Test")
	run("config", "user.email", "guardian@example.invalid")
	for name, content := range files {
		target := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	run("add", "--all")
	run("commit", "-m", "baseline")
	return run("rev-parse", "HEAD")
}

func TestRepositoryPatternRejectsURLAndMalformedCoordinates(t *testing.T) {
	for _, value := range []string{"", "owner", "owner/repo/extra", "https://github.com/owner/repo", "owner/repo?token=x"} {
		if _, err := NewGitHubClient(value, ""); err == nil {
			t.Errorf("invalid repository %q was accepted", value)
		}
	}
	if _, err := NewGitHubClient("owner/repo", ""); err != nil {
		t.Fatal(fmt.Errorf("valid repository rejected: %w", err))
	}
}
