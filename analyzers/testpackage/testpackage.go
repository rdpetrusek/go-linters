// Package testpackage enforces external test packages.
package testpackage

import (
 "errors"
 "go/parser"
 "go/token"
 "os"
 "strings"

 "golang.org/x/tools/go/analysis"
)

// Analyzer requires _test.go files to use the corresponding external package.
var Analyzer = &analysis.Analyzer{Name: "testpackage", Doc: "require tests to use an external test package", Run: run}

func run(pass *analysis.Pass) (any, error) {
 for _, file := range pass.Files {
  filename := pass.Fset.PositionFor(file.Pos(), false).Filename
  if !strings.HasSuffix(filename, "_test.go") { continue }
  source := strings.TrimSuffix(filename, "_test.go") + ".go"
  data, err := os.ReadFile(source)
  if errors.Is(err, os.ErrNotExist) {
   if !strings.HasSuffix(file.Name.Name, "_test") { pass.Reportf(file.Name.Pos(), "test files must use an external package ending in _test") }
   continue
  }
  if err != nil { return nil, err }
  base, err := parser.ParseFile(token.NewFileSet(), source, data, parser.PackageClauseOnly)
  if err != nil { return nil, err }
  if file.Name.Name != base.Name.Name+"_test" { pass.Reportf(file.Name.Pos(), "test files must use an external package ending in _test") }
 }
 return nil, nil
}
