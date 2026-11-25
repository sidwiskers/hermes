package guardian

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const maxGitHubResponse = 12_000_000

var repositoryPattern = regexp.MustCompile(`^[\w.-]+/[\w.-]+$`)

// GitHubClient is a small, bounded GitHub REST client. It deliberately avoids
// returning response bodies in errors because authenticated APIs may echo
// sensitive request metadata.
type GitHubClient struct {
	Repository string
	Token      string
	BaseURL    string
	HTTP       *http.Client
}

func NewGitHubClient(repository, token string) (*GitHubClient, error) {
	if !repositoryPattern.MatchString(repository) {
		return nil, fmt.Errorf("guardian: set GITHUB_REPOSITORY to owner/repository")
	}
	return &GitHubClient{
		Repository: repository,
		Token:      token,
		BaseURL:    "https://api.github.com",
		HTTP:       &http.Client{Timeout: 45 * time.Second},
	}, nil
}

func (client *GitHubClient) route(path string) (string, error) {
	if !strings.HasPrefix(path, "/repos/") || strings.Contains(path, "://") ||
		strings.Contains(path, "\\") || strings.Contains(path, "..") {
		return "", fmt.Errorf("guardian: repository API route required")
	}
	return strings.TrimRight(client.BaseURL, "/") + path, nil
}

// Request performs one GitHub API call. A 404 is treated as absence only when
// the caller explicitly requests missing=true.
func (client *GitHubClient) Request(path, method string, body any, missing bool) (any, error) {
	endpoint, err := client.route(path)
	if err != nil {
		return nil, err
	}
	var data io.Reader
	if body != nil {
		encoded, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			return nil, marshalErr
		}
		data = bytes.NewReader(encoded)
	}
	request, err := http.NewRequest(method, endpoint, data)
	if err != nil {
		return nil, fmt.Errorf("guardian: invalid GitHub request")
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "hermes-guardian")
	request.Header.Set("Content-Type", "application/json")
	if client.Token != "" {
		request.Header.Set("Authorization", "Bearer "+client.Token)
	}
	httpClient := client.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 45 * time.Second}
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("guardian: GitHub request failed")
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound && missing {
		return nil, nil
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("guardian: GitHub %s request failed (HTTP %d)", method, response.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, maxGitHubResponse+1))
	if err != nil || len(raw) > maxGitHubResponse {
		return nil, fmt.Errorf("guardian: GitHub response unavailable or too large")
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, nil
	}
	var result any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("guardian: GitHub returned an invalid JSON response")
	}
	return result, nil
}

func (client *GitHubClient) RepoRoute() string {
	return "/repos/" + client.Repository
}

func mapValue(value any) map[string]any {
	result, _ := value.(map[string]any)
	return result
}

func arrayValue(value any) []any {
	result, _ := value.([]any)
	return result
}

func stringValue(value any) string {
	result, _ := value.(string)
	return result
}

func boolValue(value any) bool {
	result, _ := value.(bool)
	return result
}

func intValue(value any) int {
	switch result := value.(type) {
	case float64:
		return int(result)
	case int:
		return result
	case json.Number:
		var parsed int
		_, _ = fmt.Sscan(string(result), &parsed)
		return parsed
	default:
		return 0
	}
}

func nestedString(value any, keys ...string) string {
	current := value
	for _, key := range keys {
		current = mapValue(current)[key]
	}
	return stringValue(current)
}
