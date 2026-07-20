// Command botapi-audit compares Hermes' typed source with the checked-in
// official Telegram Bot API schema inventory.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/sidwiskers/hermes/internal/botapi"
)

func main() {
	specPath := flag.String("spec", "spec/bot-api.json", "checked-in Bot API manifest")
	root := flag.String("root", ".", "Hermes module root")
	jsonOutput := flag.Bool("json", false, "write the complete report as JSON")
	allowGaps := flag.Bool("allow-gaps", false, "report parity gaps without returning a failing status")
	statusPath := flag.String("status-file", "", "optional file receiving complete or gaps")
	flag.Parse()

	schema, err := botapi.Load(*specPath)
	if err != nil {
		fatal(err)
	}
	inventory, err := botapi.ScanGo(*root, "api", "types")
	if err != nil {
		fatal(err)
	}
	report := botapi.Audit(schema, inventory)
	failed := report.GapCount() != 0
	if *statusPath != "" {
		status := "complete\n"
		if failed {
			status = "gaps\n"
		}
		if err := os.WriteFile(*statusPath, []byte(status), 0o644); err != nil {
			fatal(fmt.Errorf("write audit status: %w", err))
		}
	}
	if *jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			fatal(err)
		}
		if failed && !*allowGaps {
			os.Exit(1)
		}
		return
	}
	fmt.Printf("Bot API %s: %d methods, %d parameters, %d objects, %d fields, %d unions, %d variants\n",
		schema.Version, schema.Stats().Methods, schema.Stats().Parameters, schema.Stats().Objects,
		schema.Stats().ObjectFields, schema.Stats().Unions, schema.Stats().Variants)
	fmt.Printf("parity gaps: %d\n", report.GapCount())
	fmt.Printf("  missing method bindings: %d\n", len(report.MissingMethodBindings))
	fmt.Printf("  missing method parameters: %d\n", len(report.MissingMethodParams))
	fmt.Printf("  mismatched method optionality: %d\n", len(report.MismatchedMethodOptionality))
	fmt.Printf("  mismatched method types: %d\n", len(report.MismatchedMethodTypes))
	fmt.Printf("  non-nilable optional method objects: %d\n", len(report.NonNilableMethodOptionals))
	fmt.Printf("  missing object types: %d\n", len(report.MissingObjectTypes))
	fmt.Printf("  missing object fields: %d\n", len(report.MissingObjectFields))
	fmt.Printf("  mismatched object optionality: %d\n", len(report.MismatchedObjectOptionality))
	fmt.Printf("  mismatched object types: %d\n", len(report.MismatchedObjectTypes))
	fmt.Printf("  non-nilable optional object fields: %d\n", len(report.NonNilableObjectOptionals))
	fmt.Printf("  missing union types: %d\n", len(report.MissingUnionTypes))
	fmt.Printf("  missing union variants: %d\n", len(report.MissingUnionVariants))
	if failed && !*allowGaps {
		os.Exit(1)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
