// Package plugin registers the shared custom analyzers with golangci-lint.
package plugin

import (
 "go/ast"
 "strings"

 "github.com/golangci/plugin-module-register/register"
 "golang.org/x/tools/go/analysis"
)

func init() { register.Plugin("shared", New) }

type Plugin struct{}

func New(_ any) (register.LinterPlugin, error) { return &Plugin{}, nil }
func (*Plugin) GetLoadMode() string { return register.LoadModeSyntax }
func (*Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
 return []*analysis.Analyzer{{Name: "todoowner", Doc: "report TODO comments without an owner", Run: run}}, nil
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
