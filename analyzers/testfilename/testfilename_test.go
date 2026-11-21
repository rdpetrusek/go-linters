package testfilename_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/rdpetrusek/go-linters/analyzers/testfilename"
	"golang.org/x/tools/go/analysis"
)

func TestRun(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "paired.go"), []byte("package sample"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		want int
	}{{"paired_test.go", 0}, {"orphan_test.go", 1}, {"paired.go", 0}} {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, filepath.Join(dir, tc.name), "package sample_test", 0)
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		_, err = testfilename.Analyzer.Run(&analysis.Pass{Fset: fset, Files: []*ast.File{f}, Report: func(analysis.Diagnostic) { count++ }})
		if err != nil || count != tc.want {
			t.Errorf("%s: got %d, %v; want %d", tc.name, count, err, tc.want)
		}
	}
}
