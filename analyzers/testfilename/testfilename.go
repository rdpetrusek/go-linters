// Package testfilename checks that test files have a neighboring source file.
package testfilename

import (
	"errors"
	"os"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer checks that foo_test.go is next to foo.go.
var Analyzer = &analysis.Analyzer{Name: "testfilename", Doc: "require each test file to have a corresponding source file", Run: run}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		filename := pass.Fset.PositionFor(file.Pos(), false).Filename
		if !strings.HasSuffix(filename, "_test.go") {
			continue
		}
		source := strings.TrimSuffix(filename, "_test.go") + ".go"
		_, err := os.ReadFile(source) // #nosec G304 -- source is the adjacent file for the test being analyzed.
		if errors.Is(err, os.ErrNotExist) {
			pass.Reportf(file.Name.Pos(), "test file must be beside its corresponding source file %s", source)
			continue
		}
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}
