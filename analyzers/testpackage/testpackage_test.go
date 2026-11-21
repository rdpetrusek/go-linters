package testpackage_test

import (
 "go/ast"
 "go/parser"
 "go/token"
 "testing"

 "github.com/rdpetrusek/go-linters/analyzers/testpackage"
 "golang.org/x/tools/go/analysis"
)

func TestRun(t *testing.T) {
 for _, tc := range []struct{file, source string; want int}{
  {"one_test.go", "package sample", 1},
  {"one_test.go", "package sample_test", 0},
  {"one.go", "package sample", 0},
 } {
  fset := token.NewFileSet()
  f, err := parser.ParseFile(fset, tc.file, tc.source, 0)
  if err != nil { t.Fatal(err) }
  count := 0
  _, err = testpackage.Analyzer.Run(&analysis.Pass{Fset: fset, Files: []*ast.File{f}, Report: func(analysis.Diagnostic) { count++ }})
  if err != nil || count != tc.want { t.Errorf("%s %s: got %d, %v; want %d", tc.file, tc.source, count, err, tc.want) }
 }
}
