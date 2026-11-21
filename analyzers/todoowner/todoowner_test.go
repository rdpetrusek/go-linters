package todoowner_test

import (
 "go/ast"
 "go/parser"
 "go/token"
 "testing"

 "golang.org/x/tools/go/analysis"
 "github.com/rdpetrusek/go-linters/analyzers/todoowner"
)

func TestTODOOwner(t *testing.T) {
 fset := token.NewFileSet()
 file, err := parser.ParseFile(fset, "sample.go", "package sample\n// TODO: fix\n// TODO(): fix\n// TODO(alice): fix\n", parser.ParseComments)
 if err != nil { t.Fatal(err) }
 var findings []analysis.Diagnostic
 _, err = todoowner.Analyzer.Run(&analysis.Pass{Fset: fset, Files: []*ast.File{file}, Report: func(d analysis.Diagnostic) { findings = append(findings, d) }})
 if err != nil { t.Fatal(err) }
 if len(findings) != 2 { t.Fatalf("got %d findings, want 2", len(findings)) }
}
