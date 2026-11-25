package guardian

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var updateBranchPattern = regexp.MustCompile(`^automation/bot-api-[0-9]+\.[0-9]+-[a-f0-9]{12}$`)
var releaseTagPattern = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)
var commitPattern = regexp.MustCompile(`^[a-f0-9]{40}$`)

type publicationChange struct {
	Path    string
	Mode    string
	Content string
}

func (controller *Controller) Publish() error {
	bundle, err := readJSONFile(filepath.Join(controller.Out, "bundle.json"))
	if err != nil {
		return err
	}
	report := mapValue(bundle["report"])
	rawChanges := arrayValue(bundle["changes"])
	if len(report) == 0 {
		return fmt.Errorf("guardian: publication bundle has no report")
	}
	status := stringValue(report["status"])
	if status != "ready" && status != "reviewed" && status != "needs_attention" {
		return fmt.Errorf("guardian: invalid candidate status")
	}
	branch := stringValue(report["branch"])
	if !updateBranchPattern.MatchString(branch) {
		return fmt.Errorf("guardian: invalid update branch")
	}
	changes := make([]publicationChange, 0, len(rawChanges))
	totalBytes := 0
	seen := make(map[string]struct{}, len(rawChanges))
	for _, raw := range rawChanges {
		change := mapValue(raw)
		name := stringValue(change["path"])
		content, isText := change["content"].(string)
		mode := stringValue(change["mode"])
		if name == "" {
			return fmt.Errorf("guardian: publication contains an invalid file path")
		}
		if _, duplicate := seen[name]; duplicate {
			return fmt.Errorf("guardian: publication contains a duplicate path")
		}
		seen[name] = struct{}{}
		if !isGeneratedPath(name) && !AllowedRepairPath(name, false) {
			return fmt.Errorf("guardian: publication includes a protected path")
		}
		if mode != "100644" || !isText {
			return fmt.Errorf("guardian: publication must contain regular text files")
		}
		if len(content) > 2_000_000 {
			return fmt.Errorf("guardian: publication file exceeds size limit")
		}
		totalBytes += len(content)
		changes = append(changes, publicationChange{Path: name, Mode: mode, Content: content})
	}
	if totalBytes > 6_000_000 {
		return fmt.Errorf("guardian: publication exceeds size limit")
	}
	github, err := controller.github()
	if err != nil {
		return err
	}
	route := github.RepoRoute()
	mainResult, err := github.Request(route+"/git/ref/heads/main", "GET", nil, false)
	if err != nil {
		return err
	}
	mainSHA := nestedString(mainResult, "object", "sha")
	if mainSHA == "" || mainSHA != stringValue(report["base"]) {
		return fmt.Errorf("guardian: main changed during validation; rerun Guardian against the new main")
	}
	existing, err := github.Request(route+"/git/ref/heads/"+branch, "GET", nil, true)
	if err != nil {
		return err
	}
	expected := stringValue(report["expected_head"])
	actualExpected := nestedString(existing, "object", "sha")
	if (existing == nil && expected != "") || (existing != nil && actualExpected != expected) {
		return fmt.Errorf("guardian: update branch changed during validation; rerun to preserve that work")
	}
	baseCommit, err := github.Request(route+"/git/commits/"+mainSHA, "GET", nil, false)
	if err != nil {
		return err
	}
	treeEntries := make([]map[string]any, 0, len(changes))
	for _, change := range changes {
		treeEntries = append(treeEntries, map[string]any{
			"path": change.Path, "mode": change.Mode, "type": "blob", "content": change.Content,
		})
	}
	tree, err := github.Request(route+"/git/trees", "POST", map[string]any{
		"base_tree": nestedString(baseCommit, "tree", "sha"),
		"tree":      treeEntries,
	}, false)
	if err != nil {
		return err
	}
	treeSHA := nestedString(tree, "sha")
	if treeSHA == "" {
		return fmt.Errorf("guardian: GitHub did not return the candidate tree")
	}
	var oldCommit any
	if expected != "" {
		oldCommit, err = github.Request(route+"/git/commits/"+expected, "GET", nil, false)
		if err != nil {
			return err
		}
	}
	commitSHA := expected
	if oldCommit == nil || nestedString(oldCommit, "tree", "sha") != treeSHA {
		parents := []string{mainSHA}
		if expected != "" {
			parents = []string{expected}
			if expected != mainSHA {
				parents = append(parents, mainSHA)
			}
		}
		commit, commitErr := github.Request(route+"/git/commits", "POST", map[string]any{
			"message": "Maintain Telegram Bot API " + stringValue(report["version"]),
			"tree":    treeSHA,
			"parents": parents,
		}, false)
		if commitErr != nil {
			return commitErr
		}
		commitSHA = nestedString(commit, "sha")
		if commitSHA == "" {
			return fmt.Errorf("guardian: GitHub did not return the candidate commit")
		}
		if existing != nil {
			_, err = github.Request(route+"/git/refs/heads/"+branch, "PATCH",
				map[string]any{"sha": commitSHA, "force": false}, false)
		} else {
			_, err = github.Request(route+"/git/refs", "POST",
				map[string]any{"ref": "refs/heads/" + branch, "sha": commitSHA}, false)
		}
		if err != nil {
			return err
		}
	}

	owner := strings.SplitN(github.Repository, "/", 2)[0]
	query := url.Values{}
	query.Set("state", "open")
	query.Set("head", owner+":"+branch)
	query.Set("base", "main")
	prValue, err := github.Request(route+"/pulls?"+query.Encode(), "GET", nil, false)
	if err != nil {
		return err
	}
	prs := arrayValue(prValue)
	diff, _ := os.ReadFile(filepath.Join(controller.Out, "diff.md"))
	validation := mapValue(report["validation"])
	validationOutput := stringValue(validation["output"])
	if len(validationOutput) > 12_000 {
		validationOutput = validationOutput[len(validationOutput)-12_000:]
	}
	body := controller.summary(report) + "\n" + string(diff) +
		"\n### Validation\n\n\x60\x60\x60text\n" + validationOutput + "\n\x60\x60\x60\n"
	body += fmt.Sprintf("\nSource identity: `%s`\n\nValidated base: `%s`\n",
		stringValue(report["identity"]), mainSHA)
	prData := map[string]any{
		"title": "Update Hermes for Telegram Bot API " + stringValue(report["version"]),
		"body":  body,
	}
	var pr map[string]any
	if len(prs) > 0 {
		number := intValue(mapValue(prs[0])["number"])
		prValue, err = github.Request(route+"/pulls/"+fmt.Sprint(number), "PATCH", prData, false)
		if err != nil {
			return err
		}
		pr = mapValue(prValue)
	} else {
		prData["head"] = branch
		prData["base"] = "main"
		prData["draft"] = status != "ready"
		prValue, err = github.Request(route+"/pulls", "POST", prData, false)
		if err != nil {
			return err
		}
		pr = mapValue(prValue)
	}
	if len(pr) == 0 {
		return fmt.Errorf("guardian: GitHub did not return the pull request")
	}
	report["pr"] = stringValue(pr["html_url"])
	report["head"] = commitSHA
	if err := writeJSONFile(filepath.Join(controller.Out, "report.json"), report); err != nil {
		return err
	}

	// An existing draft is never silently promoted. An owner may have
	// deliberately paused it; protected-branch rules remain authoritative.
	if controller.getenv("GUARDIAN_MODE") == "automatic" && AutomaticCandidate(report, changes) && !boolValue(pr["draft"]) {
		latest, latestErr := github.Request(route+"/git/ref/heads/main", "GET", nil, false)
		if latestErr != nil {
			return latestErr
		}
		if nestedString(latest, "object", "sha") != mainSHA {
			return fmt.Errorf("guardian: main moved before merge; rerun validation")
		}
		number := intValue(pr["number"])
		merged, mergeErr := github.Request(route+"/pulls/"+fmt.Sprint(number)+"/merge", "PUT",
			map[string]any{"sha": commitSHA, "merge_method": "squash"}, false)
		if mergeErr != nil {
			message := mergeErr.Error()
			if strings.Contains(message, "HTTP 405") || strings.Contains(message, "HTTP 409") || strings.Contains(message, "HTTP 422") {
				report["status"] = "awaiting_merge"
				report["message"] = "GitHub requires checks or approvals before merging. The prepared PR is preserved; the next check will retry without replacing its commit."
				return controller.finish(report)
			}
			return mergeErr
		}
		if boolValue(mapValue(merged)["merged"]) {
			mergeSHA := stringValue(mapValue(merged)["sha"])
			actual, err := github.Request(route+"/git/commits/"+mergeSHA, "GET", nil, false)
			if err != nil {
				return err
			}
			if nestedString(actual, "tree", "sha") != treeSHA {
				return fmt.Errorf("guardian: main changed while GitHub merged the PR; revalidate current main before releasing")
			}
			report["status"] = "merged"
			report["message"] = "Verified mechanical update merged successfully."
			if controller.getenv("GUARDIAN_RELEASE") == "true" {
				if err := controller.PublishRelease(stringValue(report["release"]), mergeSHA, body); err != nil {
					return err
				}
				report["status"] = "released"
				report["message"] = "Verified mechanical update merged and released successfully."
			}
		}
	}
	if err := controller.finish(report); err != nil {
		return err
	}
	if prURL := stringValue(pr["html_url"]); prURL != "" {
		fmt.Println(prURL)
	}
	return nil
}

func AutomaticCandidate(report map[string]any, changes []publicationChange) bool {
	if stringValue(report["status"]) != "ready" || report["mechanical"] != true ||
		report["parity"] != true || !boolValue(mapValue(report["validation"])["passed"]) ||
		report["repair"] != nil {
		return false
	}
	for _, change := range changes {
		if !isGeneratedPath(change.Path) {
			return false
		}
	}
	return true
}

func (controller *Controller) PublishRelease(tag, sha, notes string) error {
	if !releaseTagPattern.MatchString(tag) || !commitPattern.MatchString(sha) {
		return fmt.Errorf("guardian: invalid release target")
	}
	github, err := controller.github()
	if err != nil {
		return err
	}
	route := github.RepoRoute()
	existing, err := github.Request(route+"/git/ref/tags/"+tag, "GET", nil, true)
	if err != nil {
		return err
	}
	target := mapValue(existing)["object"]
	for depth := 0; depth < 5; depth++ {
		object := mapValue(target)
		if len(object) == 0 || stringValue(object["type"]) != "tag" {
			break
		}
		tagSHA := stringValue(object["sha"])
		if tagSHA == "" {
			return fmt.Errorf("guardian: annotated release tag has no object ID")
		}
		annotated, err := github.Request(route+"/git/tags/"+tagSHA, "GET", nil, false)
		if err != nil {
			return err
		}
		target = mapValue(annotated)["object"]
	}
	if object := mapValue(target); len(object) > 0 && stringValue(object["sha"]) != sha {
		return fmt.Errorf("guardian: release tag already exists at a different commit; never move published versions")
	}
	if existing == nil {
		if _, err := github.Request(route+"/git/refs", "POST",
			map[string]any{"ref": "refs/tags/" + tag, "sha": sha}, false); err != nil {
			return err
		}
	}
	published, err := github.Request(route+"/releases/tags/"+tag, "GET", nil, true)
	if err != nil {
		return err
	}
	if published == nil {
		_, err = github.Request(route+"/releases", "POST", map[string]any{
			"tag_name": tag, "target_commitish": sha, "name": "Hermes " + tag,
			"body": notes, "draft": false, "prerelease": false,
		}, false)
		return err
	}
	if boolValue(mapValue(published)["draft"]) || boolValue(mapValue(published)["prerelease"]) {
		_, err = github.Request(route+"/releases/"+fmt.Sprint(intValue(mapValue(published)["id"])),
			"PATCH", map[string]any{"draft": false, "prerelease": false}, false)
	}
	return err
}

func (controller *Controller) Release() error {
	report, err := readJSONFile(filepath.Join(controller.Out, "release.json"))
	if err != nil {
		return err
	}
	if stringValue(report["status"]) != "release_verified" || !boolValue(mapValue(report["validation"])["passed"]) {
		return fmt.Errorf("guardian: missing release validation")
	}
	github, err := controller.github()
	if err != nil {
		return err
	}
	main, err := github.Request(github.RepoRoute()+"/git/ref/heads/main", "GET", nil, false)
	if err != nil {
		return err
	}
	sha := nestedString(main, "object", "sha")
	if sha != stringValue(report["base"]) {
		return fmt.Errorf("guardian: main changed after release validation; rerun")
	}
	if err := controller.PublishRelease(stringValue(report["release"]), sha, controller.summary(report)); err != nil {
		return err
	}
	report["status"] = "released"
	report["message"] = "Verified release published successfully."
	return controller.finish(report)
}

func (controller *Controller) ReleaseCheck() error {
	if err := os.RemoveAll(controller.Out); err != nil {
		return err
	}
	if err := os.MkdirAll(controller.Out, 0o755); err != nil {
		return err
	}
	if err := os.RemoveAll(controller.Work); err != nil {
		return err
	}
	if _, err := controller.runOK([]string{"git", "clone", "--no-hardlinks", controller.Root, controller.Work}, controller.Root, 180*time.Second); err != nil {
		return err
	}
	shaOutput, err := controller.runOK([]string{"git", "rev-parse", "HEAD"}, controller.Root, 30*time.Second)
	if err != nil {
		return err
	}
	changelog, err := os.ReadFile(filepath.Join(controller.Root, "CHANGELOG.md"))
	if err != nil {
		return err
	}
	versionLine := regexp.MustCompile(`(?m)^## ([0-9]+\.[0-9]+\.[0-9]+) -`).FindSubmatch(changelog)
	if len(versionLine) != 2 {
		return fmt.Errorf("guardian: no version in changelog")
	}
	result, err := controller.sandbox("./scripts/release-check.sh", true, 2100*time.Second, false)
	if err != nil {
		return err
	}
	if result.Passed {
		minimum, minErr := controller.sandbox("./scripts/check-generated.sh && go vet ./... && go test -shuffle=on -count=1 ./...",
			false, 600*time.Second, true)
		if minErr != nil {
			return minErr
		}
		result = SandboxResult{Passed: minimum.Passed, Output: result.Output + "\nGo 1.25 verification:\n" + tailText(minimum.Output, 4000)}
	}
	if !result.Passed {
		return fmt.Errorf("guardian: release validation failed:\n%s", result.Output)
	}
	report := map[string]any{
		"status": "release_verified",
		"base": strings.TrimSpace(shaOutput),
		"release": "v" + string(versionLine[1]),
		"message": "Current main passed release validation.",
		"validation": result,
	}
	if err := writeJSONFile(filepath.Join(controller.Out, "release.json"), report); err != nil {
		return err
	}
	return controller.finish(report)
}
