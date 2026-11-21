package plugin

import "github.com/rdpetrusek/go-linters/analyzers/testonly"

func init() { analyzers = append(analyzers, testonly.Analyzer) }
