package botapi

import (
	"reflect"
	"sort"
)

// NamedChanges describes added, removed, and structurally changed declarations.
// Names are sorted because manifests are validated and emitted deterministically.
type NamedChanges struct {
	Added   []string `json:"added"`
	Removed []string `json:"removed"`
	Changed []string `json:"changed"`
}

// SurfaceDiff is the machine-readable change set between two Bot API manifests.
// It intentionally ignores source-document hashes and release-date metadata so
// harmless documentation edits do not create maintenance work.
type SurfaceDiff struct {
	Changed        bool         `json:"changed"`
	Classification string       `json:"classification"`
	ReviewReasons  []string     `json:"review_reasons"`
	VersionChanged bool         `json:"version_changed"`
	FromVersion    string       `json:"from_version"`
	ToVersion      string       `json:"to_version"`
	Before         Stats        `json:"before"`
	After          Stats        `json:"after"`
	Methods        NamedChanges `json:"methods"`
	Objects        NamedChanges `json:"objects"`
	Unions         NamedChanges `json:"unions"`
}

// CompareSurface compares the protocol declarations represented by two
// manifests. Anchor and source-document changes are not protocol changes.
func CompareSurface(before, after Manifest) SurfaceDiff {
	result := SurfaceDiff{
		VersionChanged: before.Version != after.Version,
		FromVersion:    before.Version,
		ToVersion:      after.Version,
		Before:         before.Stats(),
		After:          after.Stats(),
		Methods: compareNamed(
			before.Methods,
			after.Methods,
			func(value Method) string { return value.Name },
			func(left, right Method) bool { return reflect.DeepEqual(left.Parameters, right.Parameters) },
		),
		Objects: compareNamed(
			before.Objects,
			after.Objects,
			func(value Object) string { return value.Name },
			func(left, right Object) bool { return reflect.DeepEqual(left.Fields, right.Fields) },
		),
		Unions: compareNamed(
			before.Unions,
			after.Unions,
			func(value Union) string { return value.Name },
			func(left, right Union) bool {
				return reflect.DeepEqual(left.Variants, right.Variants) &&
					reflect.DeepEqual(left.Alternatives, right.Alternatives)
			},
		),
	}
	result.Changed = result.VersionChanged || result.Methods.any() || result.Objects.any() || result.Unions.any()
	result.Classification, result.ReviewReasons = classifySurfaceChange(before, after, result)
	return result
}

func classifySurfaceChange(before, after Manifest, diff SurfaceDiff) (string, []string) {
	if !diff.Changed {
		return "unchanged", make([]string, 0)
	}

	reasons := make([]string, 0)
	declarationsChanged := diff.Methods.any() || diff.Objects.any() || diff.Unions.any()
	if diff.VersionChanged && !declarationsChanged {
		reasons = append(reasons, "version changed without structural declarations")
	}
	if diff.Methods.any() {
		reasons = append(reasons, "method bindings or method semantics changed")
	}
	if diff.Unions.any() {
		reasons = append(reasons, "union encoding or decoding semantics changed")
	}
	if len(diff.Objects.Removed) != 0 {
		reasons = append(reasons, "object declarations were removed")
	}
	if !objectChangesAreAdditive(before.Objects, after.Objects, diff.Objects.Changed) {
		reasons = append(reasons, "existing object fields changed non-additively")
	}
	if len(reasons) != 0 {
		return "review", reasons
	}
	return "mechanical", reasons
}

func objectChangesAreAdditive(before, after []Object, changed []string) bool {
	old := make(map[string]Object, len(before))
	current := make(map[string]Object, len(after))
	for _, object := range before {
		old[object.Name] = object
	}
	for _, object := range after {
		current[object.Name] = object
	}
	for _, name := range changed {
		previous, previousExists := old[name]
		next, nextExists := current[name]
		if !previousExists || !nextExists || !fieldsAreAdditive(previous.Fields, next.Fields) {
			return false
		}
	}
	return true
}

func fieldsAreAdditive(before, after []Field) bool {
	previous := make(map[string]Field, len(before))
	for _, field := range before {
		previous[field.Name] = field
	}

	current := make(map[string]Field, len(after))
	for _, field := range after {
		current[field.Name] = field
		if _, exists := previous[field.Name]; !exists && field.Required {
			return false
		}
	}
	for _, field := range before {
		if next, exists := current[field.Name]; !exists || !reflect.DeepEqual(field, next) {
			return false
		}
	}
	return true
}

func (changes NamedChanges) any() bool {
	return len(changes.Added) != 0 || len(changes.Removed) != 0 || len(changes.Changed) != 0
}

func compareNamed[T any](before, after []T, name func(T) string, equal func(T, T) bool) NamedChanges {
	old := make(map[string]T, len(before))
	current := make(map[string]T, len(after))
	for _, value := range before {
		old[name(value)] = value
	}
	for _, value := range after {
		current[name(value)] = value
	}

	result := NamedChanges{
		Added:   make([]string, 0),
		Removed: make([]string, 0),
		Changed: make([]string, 0),
	}
	for _, value := range after {
		entryName := name(value)
		previous, exists := old[entryName]
		if !exists {
			result.Added = append(result.Added, entryName)
			continue
		}
		if !equal(previous, value) {
			result.Changed = append(result.Changed, entryName)
		}
	}
	for _, value := range before {
		entryName := name(value)
		if _, exists := current[entryName]; !exists {
			result.Removed = append(result.Removed, entryName)
		}
	}
	sort.Strings(result.Added)
	sort.Strings(result.Removed)
	sort.Strings(result.Changed)
	return result
}
