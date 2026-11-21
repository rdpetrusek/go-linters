// Package tabletest checks that Go tests are table driven.
package tabletest

import (
	"go/ast"
	"go/types"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
)

// Analyzer requires each Test function to range over a table of structs.
var Analyzer = &analysis.Analyzer{Name: "tabletest", Doc: "require table-driven test functions", Run: run}

func isTable(t types.Type) bool {
	if t == nil {
		return false
	}
	var elem types.Type
	switch u := t.Underlying().(type) {
	case *types.Slice:
		elem = u.Elem()
	case *types.Array:
		elem = u.Elem()
	case *types.Map:
		elem = u.Elem()
	default:
		return false
	}
	if ptr, ok := elem.Underlying().(*types.Pointer); ok {
		elem = ptr.Elem()
	}
	_, ok := elem.Underlying().(*types.Struct)
	return ok
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if !strings.HasSuffix(pass.Fset.PositionFor(file.Pos(), false).Filename, "_test.go") {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Body == nil || !strings.HasPrefix(fn.Name.Name, "Test") || fn.Name.Name == "TestMain" {
				continue
			}
			suffix := strings.TrimPrefix(fn.Name.Name, "Test")
			if suffix != "" {
				first, _ := utf8.DecodeRuneInString(suffix)
				if unicode.IsLower(first) {
					continue
				}
			}
			found := false
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				if r, ok := node.(*ast.RangeStmt); ok && isTable(pass.TypesInfo.TypeOf(r.X)) {
					found = true
					return false
				}
				return !found
			})
			if !found {
				pass.Reportf(fn.Name.Pos(), "%s must be table-driven: range over a table of structs", fn.Name.Name)
			}
		}
	}
	return nil, nil
}
