package testfunctionname_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rdpetrusek/go-linters/analyzers/testfunctionname"
	"golang.org/x/tools/go/analysis"
)

func TestRun(t *testing.T) {
	dir := t.TempDir()
	source := "package sample\nfunc Foo() {}\nfunc bar() {}\ntype Thing struct{}\nfunc (Thing) Method() {}\n"
	if err := os.WriteFile(filepath.Join(dir, "sample.go"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name        string
		wantFinding bool
	}{
		{"TestFoo", false}, {"TestBar", false}, {"TestMethod", false}, {"TestMissing", true}, {"TestMain", false}, {"helper", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, filepath.Join(dir, "sample_test.go"), "package sample_test\nfunc "+tc.name+"() {}\n", 0)
			if err != nil {
				t.Fatal(err)
			}
			var findings []string
			_, err = testfunctionname.Analyzer.Run(&analysis.Pass{Fset: fset, Files: []*ast.File{f}, Report: func(d analysis.Diagnostic) { findings = append(findings, d.Message) }})
			if err != nil {
				t.Fatal(err)
			}
			if tc.wantFinding && (len(findings) != 1 || !strings.Contains(findings[0], tc.name)) {
				t.Fatalf("expected finding for %s, got %v", tc.name, findings)
			}
			if !tc.wantFinding && len(findings) != 0 {
				t.Fatalf("unexpected findings: %v", findings)
			}
		})
	}
}
