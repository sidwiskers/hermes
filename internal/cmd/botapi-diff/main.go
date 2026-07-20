// Command botapi-diff reports semantic changes between Bot API manifests.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sidwiskers/hermes/internal/botapi"
)

func main() {
	beforePath := flag.String("before", "spec/bot-api.json", "previous Bot API manifest")
	afterPath := flag.String("after", "", "candidate Bot API manifest")
	format := flag.String("format", "markdown", "output format: markdown or json")
	statusPath := flag.String("status-file", "", "optional file receiving unchanged, mechanical, or review")
	flag.Parse()

	if *afterPath == "" {
		fatal(fmt.Errorf("-after is required"))
	}
	before, err := botapi.Load(*beforePath)
	if err != nil {
		fatal(fmt.Errorf("load previous manifest: %w", err))
	}
	after, err := botapi.Load(*afterPath)
	if err != nil {
		fatal(fmt.Errorf("load candidate manifest: %w", err))
	}
	diff := botapi.CompareSurface(before, after)

	if *statusPath != "" {
		status := diff.Classification + "\n"
		if err := os.WriteFile(*statusPath, []byte(status), 0o644); err != nil {
			fatal(fmt.Errorf("write status: %w", err))
		}
	}

	switch *format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(diff); err != nil {
			fatal(err)
		}
	case "markdown":
		writeMarkdown(diff)
	default:
		fatal(fmt.Errorf("unsupported format %q", *format))
	}
}

func writeMarkdown(diff botapi.SurfaceDiff) {
	fmt.Println("## Telegram Bot API surface change")
	fmt.Println()
	fmt.Printf("- Manifest: `%s` → `%s`\n", diff.FromVersion, diff.ToVersion)
	fmt.Printf("- Semantic change detected: `%t`\n", diff.Changed)
	fmt.Printf("- Classification: `%s`\n", diff.Classification)
	for _, reason := range diff.ReviewReasons {
		fmt.Printf("- Review required: %s\n", reason)
	}
	fmt.Println()
	fmt.Println("| Surface | Before | After |")
	fmt.Println("| --- | ---: | ---: |")
	fmt.Printf("| Methods | %d | %d |\n", diff.Before.Methods, diff.After.Methods)
	fmt.Printf("| Parameters | %d | %d |\n", diff.Before.Parameters, diff.After.Parameters)
	fmt.Printf("| Objects | %d | %d |\n", diff.Before.Objects, diff.After.Objects)
	fmt.Printf("| Object fields | %d | %d |\n", diff.Before.ObjectFields, diff.After.ObjectFields)
	fmt.Printf("| Unions | %d | %d |\n", diff.Before.Unions, diff.After.Unions)
	fmt.Printf("| Union variants | %d | %d |\n", diff.Before.Variants, diff.After.Variants)
	writeChanges("Methods", diff.Methods)
	writeChanges("Objects", diff.Objects)
	writeChanges("Unions", diff.Unions)
}

func writeChanges(label string, changes botapi.NamedChanges) {
	if len(changes.Added) == 0 && len(changes.Removed) == 0 && len(changes.Changed) == 0 {
		return
	}
	fmt.Printf("\n### %s\n\n", label)
	writeNames("Added", changes.Added)
	writeNames("Removed", changes.Removed)
	writeNames("Changed", changes.Changed)
}

func writeNames(label string, names []string) {
	if len(names) == 0 {
		return
	}
	quoted := make([]string, len(names))
	for index, name := range names {
		quoted[index] = "`" + name + "`"
	}
	fmt.Printf("- %s: %s\n", label, strings.Join(quoted, ", "))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
