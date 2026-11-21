// Package plugin registers the shared custom analyzers with golangci-lint.
package plugin

import (
	"github.com/golangci/plugin-module-register/register"
	"github.com/rdpetrusek/go-linters/analyzers/multiimpl"
	"github.com/rdpetrusek/go-linters/analyzers/todoowner"
	"golang.org/x/tools/go/analysis"
)

var analyzers = []*analysis.Analyzer{todoowner.Analyzer, multiimpl.Analyzer}

func init() { register.Plugin("shared", New) }

// Plugin provides the shared analyzers to golangci-lint.
type Plugin struct{}

// New constructs the shared analyzer plugin.
func New(_ any) (register.LinterPlugin, error) { return &Plugin{}, nil }

// GetLoadMode requests type information for the analyzers.
func (*Plugin) GetLoadMode() string { return register.LoadModeTypesInfo }

// BuildAnalyzers returns the registered analyzers.
func (*Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return analyzers, nil
}
