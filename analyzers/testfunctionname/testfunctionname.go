// Package testfunctionname checks test names against the neighboring source file.
package testfunctionname

import (
 "errors"
 "go/ast"
 "go/parser"
 "go/token"
 "os"
 "strings"
 "unicode"
 "unicode/utf8"

 "golang.org/x/tools/go/analysis"
)

// Analyzer requires TestFoo in foo_test.go to match Foo or foo in foo.go.
var Analyzer = &analysis.Analyzer{Name: "testfunctionname", Doc: "require test names to match functions in the corresponding source file", Run: run}

func testName(name string) string {
 first, size := utf8.DecodeRuneInString(name)
 if size == 0 { return "Test" }
 return "Test" + string(unicode.ToUpper(first)) + name[size:]
}

func run(pass *analysis.Pass) (any, error) {
 for _, file := range pass.Files {
  filename := pass.Fset.PositionFor(file.Pos(), false).Filename
  if !strings.HasSuffix(filename, "_test.go") { continue }
  source := strings.TrimSuffix(filename, "_test.go") + ".go"
  content, err := os.ReadFile(source)
  if errors.Is(err, os.ErrNotExist) { continue } // testfilename reports this.
  if err != nil { return nil, err }
  original, err := parser.ParseFile(token.NewFileSet(), source, content, 0)
  if err != nil { return nil, err }
  names := make(map[string]bool)
  for _, decl := range original.Decls {
   if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name != "init" { names[testName(fn.Name.Name)] = true }
  }
  for _, decl := range file.Decls {
   fn, ok := decl.(*ast.FuncDecl)
   if !ok || fn.Recv != nil || !strings.HasPrefix(fn.Name.Name, "Test") || fn.Name.Name == "TestMain" { continue }
   suffix := strings.TrimPrefix(fn.Name.Name, "Test")
   if suffix != "" { first, _ := utf8.DecodeRuneInString(suffix); if unicode.IsLower(first) { continue } }
   if !names[fn.Name.Name] { pass.Reportf(fn.Name.Pos(), "%s has no matching function in %s", fn.Name.Name, source) }
  }
 }
 return nil, nil
}
