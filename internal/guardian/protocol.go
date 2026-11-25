// Package guardian contains the Go-native repository maintenance control plane.
package guardian

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/sidwiskers/hermes/internal/botapi"
)

const semanticFormat = 1

var (
	releaseHeading = regexp.MustCompile(`^[A-Z][a-z]+ [0-9]{1,2}, [0-9]{4}$`)
	versionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+$`)
)

// Section is the normalized, non-declaration evidence captured from one
// anchored section of Telegram's official Bot API documentation.
type Section struct {
	Title string     `json:"title"`
	Body  string     `json:"body"`
	Rows  [][]string `json:"rows"`
}

// Snapshot is the stable semantic view of the documentation. Presentation-only
// HTML is discarded while prose, links and table rules remain review evidence.
type Snapshot struct {
	Format        int                `json:"format"`
	Sections      map[string]Section `json:"sections"`
	Announcements map[string]string  `json:"announcements"`
}

// CanonicalJSON serializes JSON-compatible values with sorted map keys,
// compact separators and unescaped HTML characters, matching Guardian's
// previous canonical identity representation.
func CanonicalJSON(value any) ([]byte, error) {
	// Round-trip through generic JSON values so struct fields are canonicalized
	// with the same recursively sorted key order as the previous Python code.
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var normalized any
	if err := decoder.Decode(&normalized); err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(normalized); err != nil {
		return nil, err
	}
	canonical := bytes.TrimSuffix(buffer.Bytes(), []byte("\n"))
	canonical = bytes.ReplaceAll(canonical, []byte(`\\u2028`), []byte("\u2028"))
	canonical = bytes.ReplaceAll(canonical, []byte(`\\u2029`), []byte("\u2029"))
	return canonical, nil
}

// Digest returns the SHA-256 identity of a canonical JSON-compatible value.
func Digest(value any) (string, error) {
	data, err := CanonicalJSON(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// SnapshotHTML parses Telegram's documentation into normalized semantic
// evidence. It intentionally does not replace the structural Bot API parser:
// structural declarations and human-readable behavior are independent gates.
func SnapshotHTML(document []byte) (Snapshot, error) {
	parser := documentParser{
		sections: make(map[string]*sectionBuilder),
	}
	if err := parser.feed(document); err != nil {
		return Snapshot{}, err
	}

	result := Snapshot{
		Format:        semanticFormat,
		Sections:      make(map[string]Section),
		Announcements: make(map[string]string),
	}
	for anchor, source := range parser.sections {
		title := normalize(source.title)
		section := Section{
			Title: title,
			Body:  normalize(source.body),
			Rows:  source.rows,
		}
		if releaseHeading.MatchString(title) {
			result.Announcements[title] = section.Body
			continue
		}
		result.Sections[anchor] = section
	}
	if _, ok := result.Sections["sendmessage"]; !ok {
		return Snapshot{}, fmt.Errorf("guardian: official document is incomplete or its layout changed (missing sendmessage)")
	}
	if _, ok := result.Sections["update"]; !ok {
		return Snapshot{}, fmt.Errorf("guardian: official document is incomplete or its layout changed (missing update)")
	}
	return result, nil
}

// SemanticChanges compares documentation evidence independently from the
// structural manifest diff. A changed description, return contract, link or
// existing row requires review. New optional object rows are mechanical only
// when the structural parser independently classifies them as additive.
func SemanticChanges(before *Snapshot, after Snapshot, structural botapi.SurfaceDiff) []string {
	if before == nil || before.Format != semanticFormat {
		return []string{"Documentation baseline needs review"}
	}

	reasons := make([]string, 0)
	if !reflect.DeepEqual(before.Announcements, after.Announcements) &&
		structural.Classification == "unchanged" {
		reasons = append(reasons, "release-notes: announcement changed without structural declarations")
	}

	addedObjects := make(map[string]struct{}, len(structural.Objects.Added))
	for _, name := range structural.Objects.Added {
		addedObjects[strings.ToLower(name)] = struct{}{}
	}
	changedObjects := make(map[string]struct{}, len(structural.Objects.Changed))
	for _, name := range structural.Objects.Changed {
		changedObjects[strings.ToLower(name)] = struct{}{}
	}

	names := make(map[string]struct{}, len(before.Sections)+len(after.Sections))
	for name := range before.Sections {
		names[name] = struct{}{}
	}
	for name := range after.Sections {
		names[name] = struct{}{}
	}
	ordered := make([]string, 0, len(names))
	for name := range names {
		ordered = append(ordered, name)
	}
	sort.Strings(ordered)

	for _, name := range ordered {
		left, hadBefore := before.Sections[name]
		right, hasAfter := after.Sections[name]
		switch {
		case !hadBefore:
			if _, recognized := addedObjects[name]; !recognized {
				reasons = append(reasons, name+": new documentation section")
			}
			continue
		case !hasAfter:
			reasons = append(reasons, name+": documentation section removed")
			continue
		case reflect.DeepEqual(left, right):
			continue
		}

		_, structuralObjectChange := changedObjects[name]
		if structural.Classification == "mechanical" && structuralObjectChange &&
			left.Title == right.Title && left.Body == right.Body &&
			rowsAreAdditive(left.Rows, right.Rows) {
			continue
		}
		reasons = append(reasons, name+": description, return contract, or field rules changed")
	}
	return reasons
}

func rowsAreAdditive(before, after [][]string) bool {
	current := make(map[string][]string, len(after))
	for _, row := range after {
		if len(row) > 0 {
			current[row[0]] = row
		}
	}
	for _, row := range before {
		if len(row) == 0 {
			return false
		}
		next, ok := current[row[0]]
		if !ok || !reflect.DeepEqual(row, next) {
			return false
		}
	}
	return true
}

// ValidateCandidate applies Guardian's independent rollback and schema-loss
// guards after a candidate has passed normal manifest decoding and validation.
func ValidateCandidate(before, after botapi.Manifest) error {
	if after.Source != botapi.OfficialSource {
		return fmt.Errorf("guardian: unexpected schema source")
	}
	if !versionPattern.MatchString(after.Version) {
		return fmt.Errorf("guardian: invalid Bot API version")
	}
	oldVersion, err := parseVersion(before.Version)
	if err != nil {
		return fmt.Errorf("guardian: invalid baseline Bot API version: %w", err)
	}
	newVersion, err := parseVersion(after.Version)
	if err != nil {
		return fmt.Errorf("guardian: invalid candidate Bot API version: %w", err)
	}
	if compareVersion(newVersion, oldVersion) < 0 {
		return fmt.Errorf("guardian: Telegram source moved backwards; refusing stale snapshot")
	}

	categories := []struct {
		name string
		old  []string
		next []string
	}{
		{name: "methods", old: methodNames(before.Methods), next: methodNames(after.Methods)},
		{name: "objects", old: objectNames(before.Objects), next: objectNames(after.Objects)},
		{name: "unions", old: unionNames(before.Unions), next: unionNames(after.Unions)},
	}
	for _, category := range categories {
		if len(category.next)*10 < len(category.old)*9 {
			return fmt.Errorf("guardian: large schema loss in %s; inspect source/parser", category.name)
		}
		if hasDuplicate(category.next) {
			return fmt.Errorf("guardian: duplicate schema names in %s", category.name)
		}
	}
	return nil
}

type version struct {
	major int
	minor int
}

func parseVersion(value string) (version, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return version{}, fmt.Errorf("invalid version %q", value)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return version{}, err
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return version{}, err
	}
	return version{major: major, minor: minor}, nil
}

func compareVersion(left, right version) int {
	if left.major < right.major || left.major == right.major && left.minor < right.minor {
		return -1
	}
	if left == right {
		return 0
	}
	return 1
}

func methodNames(values []botapi.Method) []string {
	names := make([]string, len(values))
	for i := range values {
		names[i] = values[i].Name
	}
	return names
}

func objectNames(values []botapi.Object) []string {
	names := make([]string, len(values))
	for i := range values {
		names[i] = values[i].Name
	}
	return names
}

func unionNames(values []botapi.Union) []string {
	names := make([]string, len(values))
	for i := range values {
		names[i] = values[i].Name
	}
	return names
}

func hasDuplicate(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			return true
		}
		seen[value] = struct{}{}
	}
	return false
}

type sectionBuilder struct {
	title []string
	body  []string
	rows  [][]string
}

type documentParser struct {
	sections map[string]*sectionBuilder
	current  *sectionBuilder
	heading  bool
	table    int
	row      []string
	cell     []string
	skip     int
	rawTag   string
	err      error
}

func (p *documentParser) feed(data []byte) error {
	for len(data) > 0 {
		if p.err != nil {
			return p.err
		}
		if p.rawTag != "" {
			closing := []byte("</" + p.rawTag)
			index := bytes.Index(bytes.ToLower(data), closing)
			if index < 0 {
				p.handleData(string(data))
				break
			}
			p.handleData(string(data[:index]))
			data = data[index:]
			p.rawTag = ""
		}
		open := bytes.IndexByte(data, '<')
		if open < 0 {
			p.handleData(string(data))
			break
		}
		if open > 0 {
			p.handleData(string(data[:open]))
			data = data[open:]
		}
		if bytes.HasPrefix(data, []byte("<!--")) {
			end := bytes.Index(data[4:], []byte("-->"))
			if end < 0 {
				break
			}
			data = data[4+end+3:]
			continue
		}
		end, ok := findTagEnd(data)
		if !ok {
			// A malformed trailing '<' is text in HTMLParser; retaining it is
			// safer than silently deleting potentially meaningful prose.
			p.handleData(string(data))
			break
		}
		raw := data[:end]
		data = data[end:]
		token, ok := parseHTMLTag(raw)
		if !ok {
			continue
		}
		if token.closing {
			p.handleEndTag(token.name)
		} else {
			p.handleStartTag(token.name, token.attrs)
			if token.selfClosing {
				p.handleEndTag(token.name)
			}
		}
	}
	return p.err
}

type htmlTag struct {
	name        string
	attrs       map[string]string
	closing     bool
	selfClosing bool
}

func findTagEnd(data []byte) (int, bool) {
	var quote byte
	for i, value := range data {
		if quote != 0 {
			if value == quote {
				quote = 0
			}
			continue
		}
		switch value {
		case '"', '\'':
			quote = value
		case '>':
			return i + 1, true
		}
	}
	return 0, false
}

func parseHTMLTag(raw []byte) (htmlTag, bool) {
	if len(raw) < 3 || raw[0] != '<' || raw[len(raw)-1] != '>' {
		return htmlTag{}, false
	}
	content := strings.TrimSpace(string(raw[1 : len(raw)-1]))
	if content == "" || strings.HasPrefix(content, "!") || strings.HasPrefix(content, "?") {
		return htmlTag{}, false
	}
	token := htmlTag{attrs: make(map[string]string)}
	if strings.HasPrefix(content, "/") {
		token.closing = true
		content = strings.TrimSpace(content[1:])
	}
	if strings.HasSuffix(content, "/") {
		token.selfClosing = true
		content = strings.TrimSpace(strings.TrimSuffix(content, "/"))
	}
	nameEnd := 0
	for nameEnd < len(content) && !isHTMLSpace(content[nameEnd]) && content[nameEnd] != '/' {
		nameEnd++
	}
	if nameEnd == 0 {
		return htmlTag{}, false
	}
	token.name = strings.ToLower(content[:nameEnd])
	if token.closing {
		return token, true
	}
	rest := content[nameEnd:]
	for len(rest) > 0 {
		rest = strings.TrimLeftFunc(rest, isHTMLSpaceRune)
		if rest == "" || strings.HasPrefix(rest, "/") {
			break
		}
		keyEnd := 0
		for keyEnd < len(rest) && !isHTMLSpace(rest[keyEnd]) && rest[keyEnd] != '=' && rest[keyEnd] != '/' {
			keyEnd++
		}
		if keyEnd == 0 {
			rest = rest[1:]
			continue
		}
		key := strings.ToLower(rest[:keyEnd])
		rest = rest[keyEnd:]
		rest = strings.TrimLeftFunc(rest, isHTMLSpaceRune)
		value := ""
		if strings.HasPrefix(rest, "=") {
			rest = strings.TrimLeftFunc(rest[1:], isHTMLSpaceRune)
			if len(rest) > 0 && (rest[0] == '"' || rest[0] == '\'') {
				quote := rest[0]
				rest = rest[1:]
				end := strings.IndexByte(rest, quote)
				if end < 0 {
					value, rest = rest, ""
				} else {
					value, rest = rest[:end], rest[end+1:]
				}
			} else {
				end := strings.IndexFunc(rest, isHTMLSpaceRune)
				if end < 0 {
					value, rest = rest, ""
				} else {
					value, rest = rest[:end], rest[end:]
				}
			}
		}
		token.attrs[key] = html.UnescapeString(value)
	}
	return token, true
}

func isHTMLSpace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\n' || value == '\r' || value == '\f'
}

func isHTMLSpaceRune(value rune) bool {
	return value == ' ' || value == '\t' || value == '\n' || value == '\r' || value == '\f'
}

func (p *documentParser) handleStartTag(tag string, attrs map[string]string) {
	if tag == "script" || tag == "style" {
		p.skip++
		p.rawTag = tag
	}
	if isHeading(tag) {
		p.current = nil
		p.heading = true
	}
	if p.heading && tag == "a" {
		if anchor := attrs["name"]; anchor != "" {
			if _, exists := p.sections[anchor]; exists {
				// Duplicate anchors are rejected by the scanner before a snapshot
				// can be used as release evidence.
				p.err = fmt.Errorf("guardian: duplicate documentation anchor %q", anchor)
				return
			}
			section := &sectionBuilder{rows: make([][]string, 0)}
			p.sections[anchor] = section
			p.current = section
		}
	}
	if p.current == nil {
		return
	}
	if tag == "table" {
		p.table++
	}
	if tag == "tr" {
		p.row = make([]string, 0)
	}
	if tag == "td" || tag == "th" {
		p.cell = make([]string, 0)
	}
	if tag == "a" && !p.heading {
		if href := attrs["href"]; href != "" {
			p.handleData(" [" + href + "] ")
		}
	}
}

func (p *documentParser) handleEndTag(tag string) {
	if tag == "script" || tag == "style" {
		p.skip--
		if p.skip < 0 {
			p.skip = 0
		}
	}
	if isHeading(tag) {
		p.heading = false
	}
	if p.current == nil {
		return
	}
	if (tag == "td" || tag == "th") && p.cell != nil {
		if p.row != nil {
			p.row = append(p.row, normalize(p.cell))
		}
		p.cell = nil
	}
	if tag == "tr" && p.row != nil {
		p.current.rows = append(p.current.rows, p.row)
		p.row = nil
	}
	if tag == "table" {
		p.table--
		if p.table < 0 {
			p.table = 0
		}
	}
}

func (p *documentParser) handleData(value string) {
	if p.current == nil || p.skip > 0 {
		return
	}
	value = html.UnescapeString(value)
	switch {
	case p.heading:
		p.current.title = append(p.current.title, value)
	case p.cell != nil:
		p.cell = append(p.cell, value)
	case p.table == 0:
		p.current.body = append(p.current.body, value)
	}
}

func isHeading(tag string) bool {
	return tag == "h1" || tag == "h2" || tag == "h3" || tag == "h4"
}

func normalize(parts []string) string {
	return strings.Join(strings.Fields(strings.Join(parts, " ")), " ")
}
