package plugin

import (
	"github.com/rdpetrusek/go-linters/analyzers/tabletest"
	"github.com/rdpetrusek/go-linters/analyzers/testfilename"
	"github.com/rdpetrusek/go-linters/analyzers/testfunctionname"
	"github.com/rdpetrusek/go-linters/analyzers/testpackage"
)

func init() {
	analyzers = append(analyzers, tabletest.Analyzer, testfilename.Analyzer, testfunctionname.Analyzer, testpackage.Analyzer)
}
