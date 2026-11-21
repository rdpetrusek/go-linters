package testonly

import (
 "go/token"
 "os"
 "path/filepath"
 "sort"
 "strings"
 "testing"

 "golang.org/x/tools/go/analysis"
 "golang.org/x/tools/go/packages"
)

func TestModuleReferences(t *testing.T) {
 root := t.TempDir()
 files := map[string]string{
  "go.mod": "module example.com/fixture\n\ngo 1.23\n",
  "library/library.go": `package library
func OnlyTest() {}
func Production() {}
func Both() {}
func NeverUsed() {}
var TestVar int
const TestConst = 1
type Record struct { TestField int; ProductionField int }
func (Record) TestMethod() {}
`,
  "library/library_test.go": `package library
func internalTest() { OnlyTest(); Both(); _ = TestVar; _ = TestConst; _ = Record{TestField: 1}; Record{}.TestMethod() }
`,
  "external/external.go": `package external
import "example.com/fixture/library"
func Use() { library.Production(); library.Both(); _ = library.Record{ProductionField: 1} }
`,
  "external/external_test.go": `package external
import "example.com/fixture/library"
func externalTest() { library.OnlyTest() }
`,
 }
 for path, content := range files {
  file := filepath.Join(root, path)
  if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil { t.Fatal(err) }
  if err := os.WriteFile(file, []byte(content), 0644); err != nil { t.Fatal(err) }
 }
 fset := token.NewFileSet()
 pkgs, err := packages.Load(&packages.Config{Dir: root, Fset: fset, Mode: packages.LoadSyntax}, "./library")
 if err != nil || len(pkgs) != 1 || len(pkgs[0].Errors) != 0 { t.Fatalf("load: %v %v", err, pkgs) }
 var findings []string
 p := pkgs[0]
 _, err = Analyzer.Run(&analysis.Pass{Fset: fset, Files: p.Syntax, Pkg: p.Types, TypesInfo: p.TypesInfo, Report: func(d analysis.Diagnostic) { findings = append(findings, d.Message) }})
 if err != nil { t.Fatal(err) }
 sort.Strings(findings)
 want := []string{"OnlyTest is used only in tests", "TestConst is used only in tests", "TestField is used only in tests", "TestMethod is used only in tests", "TestVar is used only in tests"}
 if strings.Join(findings, "|") != strings.Join(want, "|") { t.Fatalf("got %q, want %q", findings, want) }
}
