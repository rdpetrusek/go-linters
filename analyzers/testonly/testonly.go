// Package testonly finds production declarations referenced only from tests.
package testonly

import (
 "fmt"
 "go/ast"
 "go/token"
 "go/types"
 "os"
 "path/filepath"
 "strings"
 "sync"

 "golang.org/x/tools/go/analysis"
 "golang.org/x/tools/go/packages"
)

// Analyzer reports production declarations used in tests but not production code.
var Analyzer = &analysis.Analyzer{Name: "testonly", Doc: "report production declarations used only in tests", Run: run}

type position struct { file string; offset int }
type uses struct { production, test bool }
type moduleIndex struct { once sync.Once; uses map[position]uses; err error }
var modules sync.Map

func key(fset *token.FileSet, pos token.Pos) (position, bool) {
 if !pos.IsValid() { return position{}, false }
 p := fset.PositionFor(pos, false)
 if p.Filename == "" { return position{}, false }
 return position{filepath.Clean(p.Filename), p.Offset}, true
}

func rootOf(filename string) (string, error) {
 dir := filepath.Dir(filename)
 for {
  _, err := os.Stat(filepath.Join(dir, "go.mod"))
  if err == nil { return dir, nil }
  if !os.IsNotExist(err) { return "", err }
  parent := filepath.Dir(dir)
  if dir == parent { return "", fmt.Errorf("no go.mod found for %s", filename) }
  dir = parent
 }
}

func load(root string) (map[position]uses, error) {
 raw, _ := modules.LoadOrStore(root, &moduleIndex{})
 index := raw.(*moduleIndex)
 index.once.Do(func() {
  fset := token.NewFileSet()
  cfg := &packages.Config{Dir: root, Fset: fset, Tests: true, Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedDeps | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedModule}
  pkgs, err := packages.Load(cfg, "./...")
  if err != nil { index.err = err; return }
  index.uses = make(map[position]uses)
  for _, pkg := range pkgs {
   if pkg.Module == nil || filepath.Clean(pkg.Module.Dir) != filepath.Clean(root) || strings.HasSuffix(pkg.ID, ".test") { continue }
   if len(pkg.Errors) > 0 { index.err = fmt.Errorf("cannot load %s: %s", pkg.ID, pkg.Errors[0]); return }
   if pkg.TypesInfo == nil { continue }
   for id, obj := range pkg.TypesInfo.Uses {
    if obj == nil || obj.Pkg() == nil { continue }
    decl, ok := key(fset, obj.Pos())
    if !ok { continue }
    filename := fset.PositionFor(id.Pos(), false).Filename
    if filename == "" { continue }
    v := index.uses[decl]
    if strings.HasSuffix(filename, "_test.go") { v.test = true } else { v.production = true }
    index.uses[decl] = v
   }
  }
 })
 return index.uses, index.err
}

func run(pass *analysis.Pass) (any, error) {
 for _, file := range pass.Files {
  filename := pass.Fset.PositionFor(file.Pos(), false).Filename
  if strings.HasSuffix(filename, "_test.go") { continue }
  root, err := rootOf(filename)
  if err != nil { return nil, err }
  usage, err := load(root)
  if err != nil { return nil, err }
  report := func(id *ast.Ident) {
   if id == nil || id.Name == "_" { return }
   obj := pass.TypesInfo.Defs[id]
   if obj == nil { return }
   pos, ok := key(pass.Fset, obj.Pos())
   if !ok { return }
   if v := usage[pos]; v.test && !v.production { pass.Reportf(id.Pos(), "%s is used only in tests", id.Name) }
  }
  for _, decl := range file.Decls {
   switch d := decl.(type) {
   case *ast.FuncDecl:
    report(d.Name)
   case *ast.GenDecl:
    for _, spec := range d.Specs {
     switch s := spec.(type) {
     case *ast.ValueSpec:
      for _, name := range s.Names { report(name) }
     case *ast.TypeSpec:
      ast.Inspect(s.Type, func(node ast.Node) bool {
       st, ok := node.(*ast.StructType)
       if !ok { return true }
       for _, field := range st.Fields.List { for _, name := range field.Names { report(name) } }
       return true
      })
     }
    }
   }
  }
 }
 return nil, nil
}
