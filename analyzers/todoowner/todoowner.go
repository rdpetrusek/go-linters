// Package todoowner provides an analyzer for TODO comments without an owner.
package todoowner

import (
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer reports TODO comments written as // TODO: or // TODO():.
var Analyzer = &analysis.Analyzer{
	Name: "todoowner",
	Doc:  "report TODO comments without an owner",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for _, group := range file.Comments {
			for _, comment := range group.List {
				if strings.HasPrefix(comment.Text, "// TODO:") || strings.HasPrefix(comment.Text, "// TODO():") {
					pass.Reportf(comment.Pos(), "TODO comment must name an owner")
				}
			}
		}
	}
	return nil, nil
}
