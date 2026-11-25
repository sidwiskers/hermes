package guardian

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAllowedRepairPath(t *testing.T) {
	cases := []struct {
		name     string
		existing bool
		want     bool
	}{
		{"api/method.go", true, true},
		{"router.go", true, true},
		{"docs/change.md", true, true},
		{"integration/live_test.go", false, false},
		{"integration/live_test.go", true, false},
		{"api/method_test.go", true, false},
		{"api/zz_generated.go", false, false},
		{"api/doc.go", false, false},
		{"internal/cmd/main.go", false, false},
		{".github/workflows/ci.yml", false, false},
		{"../escape.go", false, false},
		{"/tmp/escape.go", false, false},
	}
	for _, test := range cases {
		if got := AllowedRepairPath(test.name, test.existing); got != test.want {
			t.Errorf("AllowedRepairPath(%q, %t) = %t, want %t", test.name, test.existing, got, test.want)
		}
	}
}

func TestApplyRepairEditsExactAndAtomicValidation(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "api"), 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "api", "method.go")
	if err := os.WriteFile(file, []byte("package api\n// old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old, newText := "// old", "// new"
	content := "package api\n"
	names, err := ApplyRepairEdits(root, []FileEdit{
		{Path: "api/method.go", Old: &old, New: &newText},
		{Path: "api/new_test.go", Content: &content},
	}, map[string]struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 || !strings.Contains(string(mustRead(t, file)), "// new") {
		t.Fatalf("unexpected edits: %v", names)
	}

	badOld := "missing"
	badNew := "replacement"
	if _, err := ApplyRepairEdits(root, []FileEdit{
		{Path: "api/method.go", Old: &newText, New: &old},
		{Path: "api/method.go", Old: &badOld, New: &badNew},
	}, map[string]struct{}{}); err == nil {
		t.Fatal("ambiguous batch succeeded")
	}
	if !strings.Contains(string(mustRead(t, file)), "// new") {
		t.Fatal("failed batch partially changed a file")
	}
}

func TestApplyRepairEditsProtectsExistingTestsAndControlPaths(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"api/existing_test.go", "scripts/verify.sh", ".github/workflows/ci.yml", "go.mod"} {
		target := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
			t.Fatal(err)
		}
		old, next := "old", "new"
		_, err := ApplyRepairEdits(root, []FileEdit{{Path: name, Old: &old, New: &next}}, nil)
		if err == nil {
			t.Errorf("path %q was not rejected", name)
		}
	}
}

func TestSafeRepairPathRejectsTraversalAndSymlinks(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"../escape.go", "/tmp/escape.go", ".git/config"} {
		if _, err := SafeRepairPath(root, name); err == nil {
			t.Errorf("path %q accepted", name)
		}
	}
	target := filepath.Join(root, "api")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "real.go"), []byte("package api"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(target, "real.go"), filepath.Join(target, "link.go")); err != nil {
		t.Fatal(err)
	}
	if _, err := SafeRepairPath(root, "api/link.go"); err == nil {
		t.Fatal("symlink was accepted")
	}
}

func TestProvidersRequireHTTPSBoundedAndConfiguredKeys(t *testing.T) {
	config := []ProviderConfig{{URL: "https://provider.example/v1/chat/completions", Model: "code", KeyEnv: "GUARDIAN_KEY_1"}}
	raw, _ := json.Marshal(config)
	getenv := func(key string) string {
		switch key {
		case "GUARDIAN_PROVIDERS":
			return string(raw)
		case "GUARDIAN_KEY_1":
			return "secret"
		default:
			return ""
		}
	}
	providers, err := ProvidersFromEnvironment(getenv)
	if err != nil || len(providers) != 1 {
		t.Fatalf("providers = %#v, err=%v", providers, err)
	}
	for _, endpoint := range []string{"http://provider.example/api", "https://user:pass@provider.example/api", "https://provider.example/api?key=x"} {
		invalid, _ := json.Marshal([]ProviderConfig{{URL: endpoint, Model: "code", KeyEnv: "GUARDIAN_KEY_1"}})
		_, err := ProvidersFromEnvironment(func(key string) string {
			if key == "GUARDIAN_PROVIDERS" { return string(invalid) }
			return "secret"
		})
		if err == nil {
			t.Errorf("endpoint %q was accepted", endpoint)
		}
	}
}

func TestCompletionRejectsRedirectAndDoesNotExposeResponseBody(t *testing.T) {
	var redirected bool
	destination := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirected = true
	}))
	defer destination.Close()
	source := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL, http.StatusFound)
	}))
	defer source.Close()
	previousFactory := providerHTTPClientFactory
	providerHTTPClientFactory = source.Client
	t.Cleanup(func() { providerHTTPClientFactory = previousFactory })
	provider := ProviderConfig{URL: source.URL, Model: "code", KeyEnv: "GUARDIAN_KEY_1"}
	_, err := Completion(provider, []chatMessage{{Role: "user", Content: "test"}}, func(key string) string {
		if key == "GUARDIAN_KEY_1" {
			return "secret"
		}
		return ""
	})
	if err == nil || redirected {
		t.Fatalf("redirect was followed: err=%v reached=%v", err, redirected)
	}
}

func TestCompletionParsesJSONObjectAndRejectsInvalidCompletion(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Authorization"), "secret") {
			t.Error("authorization header missing")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{
			"message": map[string]any{"content": "```json\n{\"action\":\"done\",\"summary\":\"ok\"}\n```"},
		}}})
	}))
	defer server.Close()
	previousFactory := providerHTTPClientFactory
	providerHTTPClientFactory = server.Client
	t.Cleanup(func() { providerHTTPClientFactory = previousFactory })
	provider := ProviderConfig{URL: server.URL, Model: "code", KeyEnv: "GUARDIAN_KEY_1"}
	result, err := Completion(provider, nil, func(string) string { return "secret" })
	if err != nil || result["action"] != "done" {
		t.Fatalf("completion = %#v, err=%v", result, err)
	}
}

func TestRepairLoopFallbackAndLimits(t *testing.T) {
	var calls int
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{
			"message": map[string]any{"content": `{"action":"done","summary":"finished"}`},
		}}})
	}))
	defer server.Close()
	previousFactory := providerHTTPClientFactory
	providerHTTPClientFactory = server.Client
	t.Cleanup(func() { providerHTTPClientFactory = previousFactory })
	raw, _ := json.Marshal([]ProviderConfig{{URL: server.URL, Model: "code", KeyEnv: "GUARDIAN_KEY_1"}})
	getenv := func(key string) string {
		if key == "GUARDIAN_PROVIDERS" {
			return string(raw)
		}
		if key == "GUARDIAN_KEY_1" {
			return "secret"
		}
		return ""
	}
	result := Repair(t.TempDir(), map[string]any{}, nil, 2, nil, getenv)
	if result.Status != "proposed" || calls != 1 {
		t.Fatalf("repair = %#v, calls=%d", result, calls)
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
