package guardian

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGitHubRouteAndRepositoryValidation(t *testing.T) {
	if _, err := NewGitHubClient("bad", "secret"); err == nil {
		t.Fatal("invalid repository accepted")
	}
	client, err := NewGitHubClient("sidwiskers/hermes", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if client.RepoRoute() != "/repos/sidwiskers/hermes" {
		t.Fatalf("repo route = %q", client.RepoRoute())
	}
	for _, path := range []string{"https://evil.invalid/repos/a/b", "/users/a", "/repos/a/../b"} {
		if _, err := client.route(path); err == nil {
			t.Errorf("unsafe API route %q accepted", path)
		}
	}
}

func TestGitHubRequestHandlesMissingAndSanitizesErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Error("GitHub authorization header missing")
		}
		switch r.URL.Path {
		case "/repos/sidwiskers/hermes/missing":
			http.Error(w, "private body with secret", http.StatusNotFound)
		case "/repos/sidwiskers/hermes/failure":
			http.Error(w, "private body with secret", http.StatusUnprocessableEntity)
		default:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
		}
	}))
	defer server.Close()
	client, _ := NewGitHubClient("sidwiskers/hermes", "secret")
	client.BaseURL = server.URL

	value, err := client.Request(client.RepoRoute()+"/missing", http.MethodGet, nil, true)
	if err != nil || value != nil {
		t.Fatalf("missing response = %#v, err=%v", value, err)
	}
	_, err = client.Request(client.RepoRoute()+"/failure", http.MethodGet, nil, false)
	if err == nil || strings.Contains(err.Error(), "private body") || strings.Contains(err.Error(), "secret") {
		t.Fatalf("GitHub error leaked response details: %v", err)
	}
	value, err = client.Request(client.RepoRoute()+"/ok", http.MethodGet, nil, false)
	if err != nil || mapValue(value)["ok"] != true {
		t.Fatalf("normal response = %#v, err=%v", value, err)
	}
}
