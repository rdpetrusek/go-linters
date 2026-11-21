package multiimpl

import (
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"
)

func TestImplementationsAcrossPackages(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"go.mod": "module example.com/check\n\ngo 1.23\n",
		"contract/contract.go": `package contract
  type One interface { One() }
  type Two interface { Two() }
  type None interface { None() }
  type a struct{}
  func (a) One() {}
  type b struct{}
  func (b) Two() {}
  type Alias = b
  type Constraint interface { ~int }
  `,
		"impl/impl.go": `package impl
  type External struct{}
  func (*External) Two() {}
  `,
		"impl/ignored_test.go": `package impl
  type OnlyTest struct{}
  func (OnlyTest) One() {}
  `,
	}
	for path, content := range files {
		filename := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	pkgs, err := loadModule(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		want    int
		checked bool
	}{
		{"One", 1, true}, {"Two", 2, true}, {"None", 0, true}, {"Constraint", 0, false},
	} {
		n, checked := implementations(pkgs, "example.com/check/contract", tc.name)
		if n != tc.want || checked != tc.checked {
			t.Errorf("%s: got (%d, %t), want (%d, %t)", tc.name, n, checked, tc.want, tc.checked)
		}
	}
}

func TestAnalyzerReportsMissingImplementation(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/fixture\n\ngo 1.23\n"), 0644); err != nil {
		t.Fatal(err)
	}
	source := "package fixture\ntype Worker interface { Work() }\ntype Only struct{}\nfunc (Only) Work() {}\n"
	if err := os.WriteFile(filepath.Join(root, "fixture.go"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	pkgs, err := packages.Load(&packages.Config{Dir: root, Fset: fset, Mode: packages.LoadSyntax}, ".")
	if err != nil || len(pkgs) != 1 {
		t.Fatalf("load: %v", err)
	}
	var findings []analysis.Diagnostic
	p := pkgs[0]
	_, err = Analyzer.Run(&analysis.Pass{Fset: fset, Files: p.Syntax, Pkg: p.Types, TypesInfo: p.TypesInfo, Report: func(d analysis.Diagnostic) { findings = append(findings, d) }})
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1", len(findings))
	}
}
