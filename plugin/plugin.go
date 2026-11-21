// Package plugin registers the shared custom analyzers with golangci-lint.
package plugin

import (
 "github.com/golangci/plugin-module-register/register"
 "github.com/rdpetrusek/go-linters/analyzers/todoowner"
 "golang.org/x/tools/go/analysis"
)

func init() { register.Plugin("shared", New) }

type Plugin struct{}

func New(_ any) (register.LinterPlugin, error) { return &Plugin{}, nil }
func (*Plugin) GetLoadMode() string { return register.LoadModeSyntax }
func (*Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
 return []*analysis.Analyzer{todoowner.Analyzer}, nil
}
