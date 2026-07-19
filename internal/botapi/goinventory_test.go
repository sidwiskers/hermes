package botapi

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSchemaParityRatchet(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "..")
	schema, err := Load(filepath.Join(root, "spec", "bot-api.json"))
	if err != nil {
		t.Fatal(err)
	}
	inventory, err := ScanGo(root, "api", "types")
	if err != nil {
		t.Fatal(err)
	}
	report := Audit(schema, inventory)
	if report.GapCount() != 0 {
		t.Fatalf("Bot API schema parity regressed: %#v", report)
	}
}

func TestScanGoResolvesAliasesAndEmbeddedFields(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	directory := filepath.Join(root, "api")
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	source := `package api

type SharedParams struct {
	UserID int64 ` + "`json:\"user_id\"`" + `
}

type RequestParams struct {
	SharedParams
	Limit int ` + "`json:\"limit,omitempty\"`" + `
}

type AliasParams = RequestParams

func call(params AliasParams) {
	_ = "getThings"
}
`
	if err := os.WriteFile(filepath.Join(directory, "api.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	inventory, err := ScanGo(root, "api")
	if err != nil {
		t.Fatal(err)
	}
	fields := make(map[string]struct{})
	inventory.addTypeFields(fields, NormalizeName("AliasParams"), make(map[string]bool))
	want := map[string]struct{}{"limit": {}, "user_id": {}}
	if !reflect.DeepEqual(fields, want) {
		t.Fatalf("fields = %#v, want %#v", fields, want)
	}
}

func TestAuditFindsSchemaGaps(t *testing.T) {
	t.Parallel()
	schema := Manifest{
		Methods: []Method{{Name: "getThing", Parameters: []Field{
			{Name: "thing_id", Type: "Integer", Required: true},
			{Name: "missing", Type: "String"},
		}}},
		Objects: []Object{{Name: "Thing", Fields: []Field{
			{Name: "name", Type: "String", Required: true},
			{Name: "missing", Type: "String"},
		}}},
		Unions: []Union{{Name: "Choice", Variants: []string{"Thing", "OtherThing"}}},
	}
	inventory := GoInventory{
		Types: map[string]*GoType{
			NormalizeName("GetThingParams"): {JSONFields: map[string]GoField{"thing_id": {Tagged: true, Type: "int64"}}},
			NormalizeName("Thing"):          {Struct: true, JSONFields: map[string]GoField{"name": {Tagged: true, Type: "string"}}},
			NormalizeName("Choice"):         {Interface: true, JSONFields: map[string]GoField{}},
		},
	}

	got := Audit(schema, inventory)
	want := AuditReport{
		MissingMethodParams:  []MissingGap{{Owner: "getThing", Name: "missing"}},
		MissingObjectFields:  []MissingGap{{Owner: "Thing", Name: "missing"}},
		MissingUnionVariants: []MissingGap{{Owner: "Choice", Name: "OtherThing"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("report = %#v, want %#v", got, want)
	}
}
