// Package plugin registers the shared custom analyzers with golangci-lint.
package plugin

import (
 "github.com/golangci/plugin-module-register/register"
 "github.com/rdpetrusek/go-linters/analyzers/todoowner"
 "github.com/rdpetrusek/go-linters/analyzers/testpackage"
 "github.com/rdpetrusek/go-linters/analyzers/testfilename"
 "github.com/rdpetrusek/go-linters/analyzers/testfunctionname"
 "github.com/rdpetrusek/go-linters/analyzers/tabletest"
 "golang.org/x/tools/go/analysis"
)

func init() { register.Plugin("shared", New) }

type Plugin struct{}

func New(_ any) (register.LinterPlugin, error) { return &Plugin{}, nil }
func (*Plugin) GetLoadMode() string { return register.LoadModeTypesInfo }
func (*Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
 return []*analysis.Analyzer{todoowner.Analyzer, testpackage.Analyzer, testfilename.Analyzer, testfunctionname.Analyzer, tabletest.Analyzer}, nil
}
