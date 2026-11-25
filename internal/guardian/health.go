package guardian

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func (controller *Controller) github() (*GitHubClient, error) {
	repository := controller.getenv("GITHUB_REPOSITORY")
	token := controller.getenv("GH_TOKEN")
	if token == "" {
		token = controller.getenv("GITHUB_TOKEN")
	}
	return NewGitHubClient(repository, token)
}

func (controller *Controller) Health() error {
	if err := os.MkdirAll(controller.Out, 0o755); err != nil {
		return err
	}
	if controller.getenv("GUARDIAN_MODE") == "paused" {
		return controller.finish(map[string]any{
			"status":  "paused",
			"message": "Maintenance is intentionally paused; no recovery dispatch was made.",
		})
	}
	github, err := controller.github()
	if err != nil {
		return err
	}
	route := github.RepoRoute()
	workflowValue, err := github.Request(route+"/actions/workflows/guardian.yml", "GET", nil, false)
	if err != nil {
		return err
	}
	workflow := mapValue(workflowValue)
	runsQuery := url.Values{}
	runsQuery.Set("status", "success")
	runsQuery.Set("per_page", "1")
	runsValue, err := github.Request(route+"/actions/workflows/guardian.yml/runs?"+runsQuery.Encode(), "GET", nil, false)
	if err != nil {
		return err
	}
	runs := arrayValue(mapValue(runsValue)["workflow_runs"])
	stale := len(runs) == 0
	if !stale {
		created := stringValue(mapValue(runs[0])["created_at"])
		date, parseErr := time.Parse(time.RFC3339, created)
		if parseErr != nil {
			return fmt.Errorf("guardian: GitHub returned an invalid workflow timestamp")
		}
		stale = time.Since(date) > 48*time.Hour
	}
	report := map[string]any{"status": "healthy", "message": "Guardian completed within the last 48 hours."}
	if stale {
		report["status"] = "needs_attention"
		report["message"] = "Guardian has not completed successfully within 48 hours."
	}
	if stringValue(workflow["state"]) != "active" {
		if _, err := github.Request(route+"/actions/workflows/guardian.yml/enable", "PUT", nil, false); err != nil {
			return err
		}
		report["message"] = stringValue(report["message"]) + " Re-enabled its schedule."
	}
	if stale {
		if _, err := github.Request(route+"/actions/workflows/guardian.yml/dispatches", "POST",
			map[string]any{"ref": "main"}, false); err != nil {
			return err
		}
	}
	if err := controller.finish(report); err != nil {
		return err
	}
	if stale {
		return controller.StatusIssue(report)
	}
	return nil
}

func (controller *Controller) Status() error {
	report, err := readJSONFile(filepath.Join(controller.Out, "report.json"))
	if err != nil {
		return err
	}
	return controller.StatusIssue(report)
}

func (controller *Controller) StatusIssue(report map[string]any) error {
	github, err := controller.github()
	if err != nil {
		return err
	}
	route := github.RepoRoute()
	query := url.Values{}
	query.Set("state", "open")
	query.Set("creator", "github-actions[bot]")
	query.Set("per_page", "100")
	issuesValue, err := github.Request(route+"/issues?"+query.Encode(), "GET", nil, false)
	if err != nil {
		return err
	}
	var issue map[string]any
	for _, value := range arrayValue(issuesValue) {
		candidate := mapValue(value)
		if stringValue(candidate["title"]) != "Hermes maintenance status" {
			continue
		}
		if candidate["pull_request"] != nil {
			continue
		}
		issue = candidate
		break
	}
	body := controller.summary(report)
	if runID := controller.getenv("GITHUB_RUN_ID"); runID != "" {
		repository := controller.getenv("GITHUB_REPOSITORY")
		body += "\n[Open maintenance run](https://github.com/" + repository + "/actions/runs/" + runID + ")\n"
	}
	data := map[string]any{"title": "Hermes maintenance status", "body": body}
	if issue != nil {
		number := intValue(issue["number"])
		_, err = github.Request(route+"/issues/"+fmt.Sprint(number), "PATCH", data, false)
		return err
	}
	_, err = github.Request(route+"/issues", "POST", data, false)
	return err
}

func (controller *Controller) SandboxSmoke() error {
	if err := os.MkdirAll(controller.Out, 0o755); err != nil {
		return err
	}
	if err := os.RemoveAll(controller.Work); err != nil {
		return err
	}
	if _, err := controller.runOK([]string{"git", "clone", "--no-hardlinks", controller.Root, controller.Work}, controller.Root, 180*time.Second); err != nil {
		return err
	}
	result, err := controller.sandbox("./scripts/check-generated.sh && go test ./internal/botapi ./api ./internal/guardian",
		false, 600*time.Second, false)
	if err != nil {
		return err
	}
	fmt.Print(result.Output)
	if !result.Passed {
		return fmt.Errorf("guardian: read-only validation sandbox smoke test failed")
	}
	return nil
}

func reportForFailure(controller *Controller, cause error) error {
	_ = os.MkdirAll(controller.Out, 0o755)
	report := map[string]any{}
	path := filepath.Join(controller.Out, "report.json")
	if _, err := os.Stat(path); err == nil {
		if loaded, readErr := readJSONFile(path); readErr == nil {
			report = loaded
		}
	}
	report["status"] = "needs_attention"
	message := cause.Error()
	if len(message) > 6000 {
		message = message[:6000]
	}
	report["message"] = message
	if err := controller.finish(report); err != nil {
		return err
	}
	return cause
}

func Execute(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: hermes-guardian <prepare|validate|publish|health|status|release-check|release|sandbox-smoke> [--source FILE] [--retry]")
		return 2
	}
	command := args[0]
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	source := flags.String("source", "", "local official HTML snapshot for offline replay")
	retry := flags.Bool("retry", false, "owner-requested extra bounded repair attempt")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "unexpected positional arguments")
		return 2
	}
	controller, err := NewController("")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	switch command {
	case "prepare":
		err = controller.Prepare(*source, *retry)
	case "validate":
		err = controller.Validate()
	case "publish":
		err = controller.Publish()
	case "health":
		err = controller.Health()
	case "status":
		err = controller.Status()
	case "release-check":
		err = controller.ReleaseCheck()
	case "release":
		err = controller.Release()
	case "sandbox-smoke":
		err = controller.SandboxSmoke()
	default:
		fmt.Fprintln(os.Stderr, "unknown Guardian command: "+command)
		return 2
	}
	if err != nil {
		reportErr := reportForFailure(controller, err)
		if reportErr != nil {
			fmt.Fprintln(os.Stderr, reportErr)
		}
		return 1
	}
	return 0
}
