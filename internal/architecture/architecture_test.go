package architecture

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const modulePath = "github.com/sidwiskers/hermes"

var layerImports = map[string]map[string]bool{
	"":                 allow("api", "framework", "internal/runtime", "types"),
	"types":            allow(),
	"api":              allow("types"),
	"framework":        allow("api", "types"),
	"internal/runtime": allow("api", "types"),
	"session":          allow("framework"),
	"fsm":              allow("framework", "session"),
	"dedupe":           allow("framework"),
	"ratelimit":        allow("framework", "session"),
	"observe":          allow("api", "framework"),
	"testkit":          allow("", "api"),
}

var publicPackages = []string{
	"",
	"api",
	"dedupe",
	"framework",
	"fsm",
	"observe",
	"ratelimit",
	"session",
	"testkit",
	"types",
}

func TestPackageDependencyDirection(t *testing.T) {
	root := repositoryRoot(t)
	for layer, permitted := range layerImports {
		layer, permitted := layer, permitted
		t.Run(displayName(layer), func(t *testing.T) {
			for _, imported := range productionImports(t, filepath.Join(root, layer)) {
				if imported == modulePath {
					if !permitted[""] {
						t.Errorf("%s imports root package; allowed module imports: %s", displayName(layer), allowedList(permitted))
					}
					continue
				}
				if strings.HasPrefix(imported, modulePath+"/") {
					dependency := strings.TrimPrefix(imported, modulePath+"/")
					if !permitted[dependency] {
						t.Errorf("%s imports %s; allowed module imports: %s", displayName(layer), dependency, allowedList(permitted))
					}
					continue
				}
				if strings.Contains(strings.SplitN(imported, "/", 2)[0], ".") {
					t.Errorf("%s imports external module %s; runtime packages must remain standard-library-only", displayName(layer), imported)
				}
			}
		})
	}
}

func TestPublicPackagesKeepDedicatedDocumentation(t *testing.T) {
	root := repositoryRoot(t)
	for _, packageDir := range publicPackages {
		doc := filepath.Join(root, packageDir, "doc.go")
		if _, err := os.Stat(doc); err != nil {
			t.Errorf("%s: dedicated package documentation: %v", displayName(packageDir), err)
		}
	}
}

func productionImports(t *testing.T, directory string) []string {
	t.Helper()

	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	var imports []string
	files := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(files, filepath.Join(directory, name), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, spec := range parsed.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("unquote import %s in %s: %v", spec.Path.Value, name, err)
			}
			imports = append(imports, path)
		}
	}
	return imports
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate architecture test")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func allow(packages ...string) map[string]bool {
	result := make(map[string]bool, len(packages))
	for _, packagePath := range packages {
		result[packagePath] = true
	}
	return result
}

func allowedList(packages map[string]bool) string {
	if len(packages) == 0 {
		return "(none)"
	}
	names := make([]string, 0, len(packages))
	for packagePath := range packages {
		names = append(names, displayName(packagePath))
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func displayName(packagePath string) string {
	if packagePath == "" {
		return "root hermes"
	}
	return packagePath
}
