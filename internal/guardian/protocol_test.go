package guardian

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/sidwiskers/hermes/internal/botapi"
)

func fixtureDocument(body, extraRows string) []byte {
	return []byte(`<h4><a name="sendmessage"></a>sendMessage</h4><p>` + body + `</p>
<table><tr><th>Parameter</th><th>Type</th><th>Required</th><th>Description</th></tr>
<tr><td>text</td><td>String</td><td>Yes</td><td>Message text</td></tr>` + extraRows + `</table>
<h4><a name="update"></a>Update</h4><p>An update.</p>`)
}

func structuralDiff(classification string) botapi.SurfaceDiff {
	return botapi.SurfaceDiff{
		Classification: classification,
		Objects: botapi.NamedChanges{
			Added:   []string{},
			Changed: []string{},
			Removed: []string{},
		},
	}
}

func TestSnapshotIgnoresHTMLPresentationAndWhitespace(t *testing.T) {
	before, err := SnapshotHTML(fixtureDocument("Returns True on success.", ""))
	if err != nil {
		t.Fatal(err)
	}
	after, err := SnapshotHTML(fixtureDocument("Returns  <b>True</b> on success.", ""))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("presentation-only HTML changed snapshot:\nbefore=%#v\nafter=%#v", before, after)
	}
}

func TestSnapshotPreservesLinkDestinations(t *testing.T) {
	first, err := SnapshotHTML(fixtureDocument(`<a href="tg://first">link</a>`, ""))
	if err != nil {
		t.Fatal(err)
	}
	second, err := SnapshotHTML(fixtureDocument(`<a href="tg://second">link</a>`, ""))
	if err != nil {
		t.Fatal(err)
	}
	if first.Sections["sendmessage"].Body == second.Sections["sendmessage"].Body {
		t.Fatal("changing an explanatory link destination was not retained")
	}
	if !strings.Contains(first.Sections["sendmessage"].Body, "[tg://first]") {
		t.Fatalf("link destination missing from body: %q", first.Sections["sendmessage"].Body)
	}
}

func TestSnapshotSkipsScriptAndStyleAndRejectsDuplicateAnchors(t *testing.T) {
	document := append(fixtureDocument("Returns True.", ""),
		[]byte(`<script>evil <a name="ignored"></a></script><style>.hidden { color: red }</style>`)...)
	snapshot, err := SnapshotHTML(document)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := snapshot.Sections["ignored"]; exists {
		t.Fatal("script content created a section")
	}
	duplicate := append(fixtureDocument("Returns True.", ""),
		[]byte(`<h4><a name="update"></a>Update</h4>`)...)
	if _, err := SnapshotHTML(duplicate); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate anchor error = %v", err)
	}
}

func TestSnapshotRequiresKnownDocumentSections(t *testing.T) {
	if _, err := SnapshotHTML([]byte("<html>temporary error</html>")); err == nil {
		t.Fatal("incomplete document was accepted")
	}
}

func TestSemanticChangesDetectsProseAndAnnouncements(t *testing.T) {
	before, err := SnapshotHTML(fixtureDocument("Returns True on success.", ""))
	if err != nil {
		t.Fatal(err)
	}
	changed, err := SnapshotHTML(fixtureDocument("Returns a Message on success.", ""))
	if err != nil {
		t.Fatal(err)
	}
	if got := SemanticChanges(&before, changed, structuralDiff("unchanged")); len(got) == 0 ||
		!strings.Contains(got[0], "sendmessage") {
		t.Fatalf("return contract change was not detected: %v", got)
	}

	announcementDoc := append(fixtureDocument("Returns True on success.", ""),
		[]byte(`<h4><a name="august-24-2026"></a>August 24, 2026</h4><p>Bot API 10.3 changed delivery behavior.</p>`)...)
	announced, err := SnapshotHTML(announcementDoc)
	if err != nil {
		t.Fatal(err)
	}
	if got := SemanticChanges(&before, announced, structuralDiff("unchanged")); len(got) == 0 ||
		!strings.Contains(got[0], "release-notes:") {
		t.Fatalf("announcement-only change was not detected: %v", got)
	}
}

func TestSemanticChangesAllowsOnlyNewOptionalObjectRows(t *testing.T) {
	before, err := SnapshotHTML(fixtureDocument("Returns True.", ""))
	if err != nil {
		t.Fatal(err)
	}
	after, err := SnapshotHTML(fixtureDocument("Returns True.", ""))
	if err != nil {
		t.Fatal(err)
	}
	// The fixture heading is an object section in the semantic document. The
	// structural parser uses the same case-insensitive anchor name.
	before.Sections["update"] = Section{
		Title: "Update",
		Body:  "An update.",
		Rows:  [][]string{{"update_id", "Integer", "Yes", "Update identifier"}},
	}
	after.Sections["update"] = Section{
		Title: "Update",
		Body:  "An update.",
		Rows: [][]string{
			{"update_id", "Integer", "Yes", "Update identifier"},
			{"business_connection_id", "String", "No", "Optional. New."},
		},
	}
	diff := structuralDiff("mechanical")
	diff.Objects.Changed = []string{"Update"}
	if got := SemanticChanges(&before, after, diff); len(got) != 0 {
		t.Fatalf("new optional object row required review: %v", got)
	}

	after.Sections["update"] = Section{
		Title: "Update",
		Body:  "An update.",
		Rows:  [][]string{{"update_id", "Integer", "Yes", "Changed rules"}},
	}
	if got := SemanticChanges(&before, after, diff); len(got) == 0 {
		t.Fatal("changed existing row was accepted as mechanical")
	}
}

func TestSemanticChangesRejectsUnknownNewSections(t *testing.T) {
	before, err := SnapshotHTML(fixtureDocument("Returns True.", ""))
	if err != nil {
		t.Fatal(err)
	}
	after := before
	after.Sections = make(map[string]Section, len(before.Sections)+1)
	for name, section := range before.Sections {
		after.Sections[name] = section
	}
	after.Sections["new-rule"] = Section{Title: "Rule", Body: "Changed", Rows: [][]string{}}
	if got := SemanticChanges(&before, after, structuralDiff("mechanical")); len(got) == 0 {
		t.Fatal("unknown new documentation section was not flagged")
	}
}

func TestSemanticChangesRequiresBaseline(t *testing.T) {
	after, err := SnapshotHTML(fixtureDocument("Returns True.", ""))
	if err != nil {
		t.Fatal(err)
	}
	if got := SemanticChanges(nil, after, structuralDiff("unchanged")); len(got) != 1 ||
		got[0] != "Documentation baseline needs review" {
		t.Fatalf("missing baseline result = %v", got)
	}
}

func TestCanonicalJSONSortsStructKeysAndDoesNotEscapeHTML(t *testing.T) {
	value := Snapshot{Format: 1, Sections: map[string]Section{}, Announcements: map[string]string{}}
	data, err := CanonicalJSON(value)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `{"announcements":{},"format":1,"sections":{}}`; got != want {
		t.Fatalf("canonical struct JSON = %s, want %s", got, want)
	}
	data, err = CanonicalJSON(map[string]string{"text": "<tag>&"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `{"text":"<tag>&"}`; got != want {
		t.Fatalf("canonical HTML escaping = %s, want %s", got, want)
	}
}

func TestCanonicalJSONPreservesUnicodeLineSeparators(t *testing.T) {
	value := map[string]string{"text": "before\u2028middle\u2029after"}
	data, err := CanonicalJSON(value)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), value["text"]) || strings.Contains(string(data), `\u2028`) || strings.Contains(string(data), `\u2029`) {
		t.Fatalf("canonical JSON escaped Unicode line separators: %s", data)
	}
}

func TestCanonicalDigestIgnoresMapInsertionOrder(t *testing.T) {
	left := map[string]any{"a": 1, "b": 2}
	right := map[string]any{"b": 2, "a": 1}
	first, err := Digest(left)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Digest(right)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("canonical digest differs: %s != %s", first, second)
	}
}

func TestValidateCandidateRejectsRollbackAndLargeSchemaLoss(t *testing.T) {
	before := validManifest(10, 1)
	after := validManifest(10, 1)
	if err := ValidateCandidate(before, after); err != nil {
		t.Fatalf("valid candidate rejected: %v", err)
	}
	after.Version = "10.0"
	if err := ValidateCandidate(before, after); err == nil {
		t.Fatal("version rollback accepted")
	}

	after = validManifest(10, 4)
	after.Methods = after.Methods[:8]
	if err := ValidateCandidate(before, after); err == nil {
		t.Fatal("10%+ method loss accepted")
	}
}

func TestValidateCandidateRejectsDuplicateNamesAndUntrustedSource(t *testing.T) {
	before := validManifest(2, 0)
	after := validManifest(2, 1)
	after.Methods[1].Name = after.Methods[0].Name
	if err := ValidateCandidate(before, after); err == nil {
		t.Fatal("duplicate method names accepted")
	}
	after = validManifest(2, 1)
	after.Source = "https://example.invalid/api"
	if err := ValidateCandidate(before, after); err == nil {
		t.Fatal("untrusted source accepted")
	}
}

func validManifest(methodCount int, minor int) botapi.Manifest {
	methods := make([]botapi.Method, methodCount)
	for i := range methods {
		methods[i] = botapi.Method{Name: "method" + strconv.Itoa(i), Anchor: "method" + strconv.Itoa(i)}
	}
	return botapi.Manifest{
		Version:      fmt.Sprintf("10.%d", minor),
		Released:     "today",
		Source:       botapi.OfficialSource,
		SourceSHA256: "sha",
		Methods:      methods,
		Objects: []botapi.Object{{
			Name: "Update", Anchor: "update",
		}},
		Unions: []botapi.Union{},
	}
}
