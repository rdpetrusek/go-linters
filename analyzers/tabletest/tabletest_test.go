package tabletest_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/rdpetrusek/go-linters/analyzers/tabletest"
	"golang.org/x/tools/go/analysis"
)

func TestRun(t *testing.T) {
	for _, tc := range []struct {
		name  string
		table types.Type
		want  int
	}{
		{"slice", types.NewSlice(types.NewStruct(nil, nil)), 0},
		{"array", types.NewArray(types.NewStruct(nil, nil), 1), 0},
		{"map", types.NewMap(types.Typ[types.String], types.NewStruct(nil, nil)), 0},
		{"nonstruct", types.NewSlice(types.Typ[types.String]), 1},
	} {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "sample_test.go", "package sample_test\nfunc TestFoo() { for range cases {} }\n", 0)
		if err != nil {
			t.Fatal(err)
		}
		info := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}
		ast.Inspect(file, func(n ast.Node) bool {
			if r, ok := n.(*ast.RangeStmt); ok {
				info.Types[r.X] = types.TypeAndValue{Type: tc.table}
			}
			return true
		})
		count := 0
		_, err = tabletest.Analyzer.Run(&analysis.Pass{Fset: fset, Files: []*ast.File{file}, TypesInfo: info, Report: func(analysis.Diagnostic) { count++ }})
		if err != nil || count != tc.want {
			t.Errorf("%s: got %d, %v; want %d", tc.name, count, err, tc.want)
		}
	}
}
