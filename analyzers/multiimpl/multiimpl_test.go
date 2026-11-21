package multiimpl_test

import (
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rdpetrusek/go-linters/analyzers/multiimpl"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"
)

func TestImplementations(t *testing.T) {
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
		if err := os.MkdirAll(filepath.Dir(filename), 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	fset := token.NewFileSet()
	pkgs, err := packages.Load(&packages.Config{Dir: root, Fset: fset, Mode: packages.LoadSyntax}, "./contract")
	if err != nil || len(pkgs) != 1 || len(pkgs[0].Errors) != 0 {
		t.Fatalf("load: %v", err)
	}
	var findings []string
	p := pkgs[0]
	_, err = multiimpl.Analyzer.Run(&analysis.Pass{Fset: fset, Files: p.Syntax, Pkg: p.Types, TypesInfo: p.TypesInfo, Report: func(d analysis.Diagnostic) { findings = append(findings, d.Message) }})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		want bool
	}{
		{"One", true}, {"Two", false}, {"None", true}, {"Constraint", false},
	} {
		found := false
		for _, finding := range findings {
			if strings.Contains(finding, "interface "+tc.name+" ") {
				found = true
			}
		}
		if found != tc.want {
			t.Errorf("%s: got %t, want %t; findings %v", tc.name, found, tc.want, findings)
		}
	}
}

func TestRun(t *testing.T) {
	cases := []struct {
		name, declarations string
		want               int
	}{
		{"none", "", 1},
		{"one", "type One struct{}; func (One) Work() {}", 1},
		{"two", "type One struct{}; func (One) Work() {}; type Two struct{}; func (*Two) Work() {}", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/fixture\n\ngo 1.23\n"), 0600); err != nil {
				t.Fatal(err)
			}
			source := "package fixture\ntype Worker interface { Work() }\n" + tc.declarations + "\n"
			if err := os.WriteFile(filepath.Join(root, "fixture.go"), []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			fset := token.NewFileSet()
			pkgs, err := packages.Load(&packages.Config{Dir: root, Fset: fset, Mode: packages.LoadSyntax}, ".")
			if err != nil || len(pkgs) != 1 {
				t.Fatalf("load: %v", err)
			}
			var findings []analysis.Diagnostic
			p := pkgs[0]
			_, err = multiimpl.Analyzer.Run(&analysis.Pass{Fset: fset, Files: p.Syntax, Pkg: p.Types, TypesInfo: p.TypesInfo, Report: func(d analysis.Diagnostic) { findings = append(findings, d) }})
			if err != nil {
				t.Fatal(err)
			}
			if len(findings) != tc.want {
				t.Fatalf("got %d findings, want %d", len(findings), tc.want)
			}
		})
	}
}
