package botapi

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseHTML(t *testing.T) {
	t.Parallel()
	data := []byte(`
<h4><a class="anchor" name="july-14-2026" href="#july-14-2026"></a>July 14, 2026</h4>
<p><strong>Bot API 10.2</strong></p>
<h4><a class="anchor" name="getthing" href="#getthing"></a>getThing</h4>
<table class="table"><thead><tr><th>Parameter</th><th>Type</th><th>Required</th><th>Description</th></tr></thead>
<tbody><tr><td>thing_id</td><td>Integer</td><td>Yes</td><td>Identifier</td></tr></tbody></table>
<h4><a class="anchor" name="thing" href="#thing"></a>Thing</h4>
<table class="table"><thead><tr><th>Field</th><th>Type</th><th>Description</th></tr></thead>
<tbody><tr><td>name</td><td>String</td><td>Optional. Name</td></tr></tbody></table>
<h4><a class="anchor" name="emptything" href="#emptything"></a>EmptyThing</h4>
<p>This object currently holds no information.</p>
<h4><a class="anchor" name="thingchoice" href="#thingchoice"></a>ThingChoice</h4>
<p>This object can be one of:</p><ul><li><a href="#thing">Thing</a></li><li><a href="#emptything">EmptyThing</a></li></ul>`)

	schema, err := ParseHTML(data)
	if err != nil {
		t.Fatal(err)
	}
	if schema.Version != "10.2" || schema.Released != "July 14, 2026" {
		t.Fatalf("metadata = %q, %q", schema.Version, schema.Released)
	}
	if stats := schema.Stats(); stats != (Stats{Methods: 1, Parameters: 1, Objects: 2, ObjectFields: 1, Unions: 1, Variants: 2}) {
		t.Fatalf("stats = %#v", stats)
	}
	if schema.Methods[0].Parameters[0] != (Field{Name: "thing_id", Type: "Integer", Required: true}) {
		t.Fatalf("parameter = %#v", schema.Methods[0].Parameters[0])
	}
	if schema.Objects[1].Fields[0] != (Field{Name: "name", Type: "String", Required: false}) {
		t.Fatalf("field = %#v", schema.Objects[1].Fields[0])
	}
	if !reflect.DeepEqual(schema.Unions[0].Variants, []string{"Thing", "EmptyThing"}) {
		t.Fatalf("variants = %#v", schema.Unions[0].Variants)
	}
}

func TestCheckedInManifest(t *testing.T) {
	t.Parallel()
	schema, err := Load(filepath.Join("..", "..", "spec", "bot-api.json"))
	if err != nil {
		t.Fatal(err)
	}
	if schema.Version != "10.2" || schema.Source != "https://core.telegram.org/bots/api" {
		t.Fatalf("metadata = %#v", schema)
	}
	want := Stats{Methods: 185, Parameters: 937, Objects: 362, ObjectFields: 1838, Unions: 26, Variants: 187}
	if got := schema.Stats(); got != want {
		t.Fatalf("stats = %#v, want %#v", got, want)
	}
}
