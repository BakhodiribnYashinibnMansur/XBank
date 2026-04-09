package arch_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBCIsolationNoDirectImports ensures that no Bounded Context
// directly imports another Bounded Context's packages.
// Communication between BCs must go through kernel or events, not direct imports.
func TestBCIsolationNoDirectImports(t *testing.T) {
	projectRoot := findProjectRoot(t)
	contextDir := filepath.Join(projectRoot, "internal", "context")

	const modulePrefix = "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/"

	bcs := discoverBCs(t, contextDir)
	if len(bcs) == 0 {
		t.Fatal("no bounded contexts found")
	}

	t.Logf("found %d bounded contexts", len(bcs))

	fset := token.NewFileSet()

	for _, bc := range bcs {
		bc := bc
		bcRel, _ := filepath.Rel(contextDir, bc.dir)
		bcImportPrefix := modulePrefix + bcRel + "/"

		t.Run(bcRel, func(t *testing.T) {
			err := filepath.Walk(bc.dir, func(path string, info os.FileInfo, err error) error {
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

					if !strings.HasPrefix(importPath, modulePrefix) {
						continue
					}
					if strings.HasPrefix(importPath, bcImportPrefix) {
						continue
					}

					relFile, _ := filepath.Rel(projectRoot, path)
					t.Errorf("%s imports another BC: %s", relFile, importPath)
				}
				return nil
			})
			if err != nil {
				t.Errorf("walking %s: %v", bcRel, err)
			}
		})
	}
}

type boundedContext struct {
	dir string
}

// discoverBCs finds all bounded contexts under internal/context/.
// Structure: internal/context/{domain}/{tier}/{bc-name}/
// Example: internal/context/banking/core/account/
func discoverBCs(t *testing.T, contextDir string) []boundedContext {
	t.Helper()

	var bcs []boundedContext

	domains, err := os.ReadDir(contextDir)
	if err != nil {
		t.Fatalf("reading context dir: %v", err)
	}

	for _, domain := range domains {
		if !domain.IsDir() {
			continue
		}
		domainPath := filepath.Join(contextDir, domain.Name())

		tiers, err := os.ReadDir(domainPath)
		if err != nil {
			continue
		}

		for _, tier := range tiers {
			if !tier.IsDir() {
				continue
			}
			tierPath := filepath.Join(domainPath, tier.Name())

			bcNames, err := os.ReadDir(tierPath)
			if err != nil {
				continue
			}

			for _, bcName := range bcNames {
				if !bcName.IsDir() {
					continue
				}
				bcs = append(bcs, boundedContext{
					dir: filepath.Join(tierPath, bcName.Name()),
				})
			}
		}
	}

	return bcs
}
