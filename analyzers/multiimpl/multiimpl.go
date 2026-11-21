// Package multiimpl checks that interfaces have multiple concrete implementations.
package multiimpl

import (
 "fmt"
 "go/ast"
 "go/types"
 "os"
 "path/filepath"
 "strings"
 "sync"

 "golang.org/x/tools/go/analysis"
 "golang.org/x/tools/go/packages"
)

// Analyzer checks implementations in non-test packages of the containing module.
var Analyzer = &analysis.Analyzer{Name: "multiimpl", Doc: "require at least two concrete implementations of each interface", Run: run}

type moduleIndex struct {
 once sync.Once
 packages []*packages.Package
 err error
}
var indexes sync.Map

func loadModule(root string) ([]*packages.Package, error) {
 item, _ := indexes.LoadOrStore(root, &moduleIndex{})
 index := item.(*moduleIndex)
 index.once.Do(func() {
  index.packages, index.err = packages.Load(&packages.Config{
   Dir: root,
   Mode: packages.NeedName | packages.NeedTypes | packages.NeedImports | packages.NeedDeps | packages.NeedModule,
  }, "./...")
  if index.err != nil { return }
  filtered := index.packages[:0]
  for _, pkg := range index.packages {
   if pkg.Module != nil && filepath.Clean(pkg.Module.Dir) == filepath.Clean(root) { filtered = append(filtered, pkg) }
  }
  index.packages = filtered
  for _, pkg := range index.packages {
   if len(pkg.Errors) != 0 { index.err = fmt.Errorf("cannot load module package %s: %s", pkg.PkgPath, pkg.Errors[0]); return }
  }
 })
 return index.packages, index.err
}

func moduleRoot(filename string) (string, error) {
 dir := filepath.Dir(filename)
 for {
  _, err := os.Stat(filepath.Join(dir, "go.mod"))
  if err == nil { return dir, nil }
  if !os.IsNotExist(err) { return "", err }
  parent := filepath.Dir(dir)
  if parent == dir { return "", fmt.Errorf("no go.mod found for %s", filename) }
  dir = parent
 }
}

func run(pass *analysis.Pass) (any, error) {
 for _, file := range pass.Files {
  filename := pass.Fset.Position(file.Pos()).Filename
  if strings.HasSuffix(filename, "_test.go") { continue }
  root, err := moduleRoot(filename)
  if err != nil { return nil, err }
  loaded, err := loadModule(root)
  if err != nil { return nil, err }
  for _, decl := range file.Decls {
   gen, ok := decl.(*ast.GenDecl)
   if !ok { continue }
   for _, spec := range gen.Specs {
    ts, ok := spec.(*ast.TypeSpec)
    if !ok || ts.Assign.IsValid() { continue }
    obj, ok := pass.TypesInfo.Defs[ts.Name].(*types.TypeName)
    if !ok { continue }
    named, ok := obj.Type().(*types.Named)
    if !ok || named.TypeParams().Len() != 0 { continue }
    if _, ok := named.Underlying().(*types.Interface); !ok { continue }
    n, checked := implementations(loaded, pass.Pkg.Path(), ts.Name.Name)
    if checked && n < 2 { pass.Reportf(ts.Name.Pos(), "interface %s has %d concrete implementations in the module; want at least 2", ts.Name.Name, n) }
   }
  }
 }
 return nil, nil
}

func implementations(pkgs []*packages.Package, targetPath, name string) (int, bool) {
 var iface *types.Interface
 for _, pkg := range pkgs {
  if pkg.PkgPath != targetPath || pkg.Types == nil { continue }
  obj, ok := pkg.Types.Scope().Lookup(name).(*types.TypeName)
  if !ok { return 0, false }
  named, ok := obj.Type().(*types.Named)
  if !ok { return 0, false }
  iface, ok = named.Underlying().(*types.Interface)
  if !ok || !iface.IsMethodSet() { return 0, false }
  iface.Complete()
  break
 }
 if iface == nil { return 0, false }
 count := 0
 for _, pkg := range pkgs {
  if pkg.Types == nil { continue }
  scope := pkg.Types.Scope()
  for _, name := range scope.Names() {
   obj, ok := scope.Lookup(name).(*types.TypeName)
   if !ok || obj.IsAlias() { continue }
   named, ok := obj.Type().(*types.Named)
   if !ok || named.TypeParams().Len() != 0 { continue }
   if _, ok := named.Underlying().(*types.Interface); ok { continue }
   if types.Implements(named, iface) || types.Implements(types.NewPointer(named), iface) { count++ }
  }
 }
 return count, true
}
