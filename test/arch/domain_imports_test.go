package arch_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDomainLayerNoExternalImports ensures that domain packages
// import only Go standard library packages.
// DDD rule: domain layer must have ZERO external dependencies.
func TestDomainLayerNoExternalImports(t *testing.T) {
	projectRoot := findProjectRoot(t)
	contextDir := filepath.Join(projectRoot, "internal", "context")

	var domainDirs []string
	err := filepath.Walk(contextDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == "domain" {
			domainDirs = append(domainDirs, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking context dir: %v", err)
	}

	if len(domainDirs) == 0 {
		t.Fatal("no domain directories found")
	}

	t.Logf("found %d domain directories", len(domainDirs))

	fset := token.NewFileSet()

	for _, domainDir := range domainDirs {
		domainDir := domainDir
		rel, _ := filepath.Rel(contextDir, domainDir)

		t.Run(rel, func(t *testing.T) {
			err := filepath.Walk(domainDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() || !strings.HasSuffix(path, ".go") {
					return nil
				}

				f, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
				if parseErr != nil {
					t.Errorf("parsing %s: %v", path, parseErr)
					return nil
				}

				for _, imp := range f.Imports {
					importPath := strings.Trim(imp.Path.Value, `"`)
					if !isAllowedDomainImport(importPath) {
						relFile, _ := filepath.Rel(projectRoot, path)
						t.Errorf("%s imports non-allowed package: %s", relFile, importPath)
					}
				}
				return nil
			})
			if err != nil {
				t.Errorf("walking %s: %v", rel, err)
			}
		})
	}
}

// Shared kernel packages that domain layers are allowed to import.
var allowedDomainImports = map[string]bool{
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain": true,
}

// isAllowedDomainImport checks if an import is stdlib or an allowed shared kernel package.
func isAllowedDomainImport(importPath string) bool {
	if allowedDomainImports[importPath] {
		return true
	}
	firstSlash := strings.Index(importPath, "/")
	firstElement := importPath
	if firstSlash > 0 {
		firstElement = importPath[:firstSlash]
	}
	return !strings.Contains(firstElement, ".")
}
