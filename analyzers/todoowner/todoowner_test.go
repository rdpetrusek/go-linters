package todoowner_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/rdpetrusek/go-linters/analyzers/todoowner"
	"golang.org/x/tools/go/analysis"
)

func TestRun(t *testing.T) {
	cases := []struct {
		name, comment string
		want          int
	}{
		{"missing owner", "// TODO: fix", 1},
		{"empty owner", "// TODO(): fix", 1},
		{"named owner", "// TODO(alice): fix", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "sample.go", "package sample\n"+tc.comment+"\n", parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			var findings []analysis.Diagnostic
			_, err = todoowner.Analyzer.Run(&analysis.Pass{Fset: fset, Files: []*ast.File{file}, Report: func(d analysis.Diagnostic) { findings = append(findings, d) }})
			if err != nil {
				t.Fatal(err)
			}
			if len(findings) != tc.want {
				t.Fatalf("got %d findings, want %d", len(findings), tc.want)
			}
		})
	}
}
