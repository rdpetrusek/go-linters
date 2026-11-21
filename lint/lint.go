// Package lint provides the shared lint runner and extension point for custom checks.
package lint

import (
 "fmt"
 "io"
 "os/exec"
)

// Check runs a lint check against packages in the current working directory.
// A check should return an error when it finds violations or cannot run.
type Check struct {
 Name string
 Run func(patterns []string, output io.Writer) error
}

// DefaultChecks is the centrally maintained baseline. Add custom checks here
// to distribute them to every consumer of a new version of the tool.
func DefaultChecks() []Check { return []Check{{Name: "vet", Run: Vet}} }

// Vet runs the Go toolchain's standard static checks.
func Vet(patterns []string, output io.Writer) error {
 args := append([]string{"vet"}, patterns...)
 cmd := exec.Command("go", args...)
 cmd.Stdout, cmd.Stderr = output, output
 return cmd.Run()
}

// Run executes every check, reporting all failures before returning an error.
func Run(checks []Check, patterns []string, output io.Writer) error {
 if len(patterns) == 0 { patterns = []string{"./..."} }
 failed := false
 for _, check := range checks {
  if check.Run == nil { fmt.Fprintf(output, "%s: missing implementation\n", check.Name); failed = true; continue }
  if err := check.Run(patterns, output); err != nil { fmt.Fprintf(output, "%s: %v\n", check.Name, err); failed = true }
 }
 if failed { return fmt.Errorf("lint checks failed") }
 return nil
}
