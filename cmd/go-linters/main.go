package main

import (
 _ "embed"
 "fmt"
 "os"
 "os/exec"
 "path/filepath"
 "runtime/debug"
 "strings"
)

const module = "github.com/rdpetrusek/go-linters"
const golangciVersion = "v2.1.6"

//go:embed config.yml
var policy []byte

func main() {
 if err := run(); err != nil { fmt.Fprintln(os.Stderr, "go-linters:", err); os.Exit(1) }
}

func run() error {
 source := os.Getenv("GO_LINTERS_SOURCE")
 info, ok := debug.ReadBuildInfo()
 if !ok { return fmt.Errorf("cannot determine module version") }
 version := info.Main.Version
 if source == "" && (version == "" || version == "(devel)") { return fmt.Errorf("use a released version or set GO_LINTERS_SOURCE to the absolute source directory") }
 dir, err := os.MkdirTemp("", "go-linters-*")
 if err != nil { return err }
 defer os.RemoveAll(dir)
 bin := filepath.Join(dir, "bin")
 if err := os.Mkdir(bin, 0755); err != nil { return err }
 install := exec.Command("go", "install", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@"+golangciVersion)
 install.Env = append(os.Environ(), "GOBIN="+bin)
 install.Stdout, install.Stderr = os.Stdout, os.Stderr
 if err := install.Run(); err != nil { return fmt.Errorf("install golangci-lint: %w", err) }
 plugin := "    version: " + version
 if source != "" {
  if !filepath.IsAbs(source) || strings.ContainsAny(source, "\r\n\"") { return fmt.Errorf("GO_LINTERS_SOURCE must be an absolute path without quotes or newlines") }
  plugin = "    path: \"" + source + "\""
 }
 build := "version: " + golangciVersion + "\nname: shared-gcl\ndestination: " + dir + "\nplugins:\n  - module: " + module + "\n    import: " + module + "/plugin\n" + plugin + "\n"
 buildConfig := filepath.Join(dir, ".custom-gcl.yml")
 if err := os.WriteFile(buildConfig, []byte(build), 0600); err != nil { return err }
 config := filepath.Join(dir, "policy.yml")
 if err := os.WriteFile(config, policy, 0600); err != nil { return err }
 custom := exec.Command(filepath.Join(bin, "golangci-lint"), "custom")
 custom.Dir = dir
 custom.Stdout, custom.Stderr = os.Stdout, os.Stderr
 if err := custom.Run(); err != nil { return fmt.Errorf("build custom golangci-lint: %w", err) }
 args := append([]string{"run", "--config", config}, os.Args[1:]...)
 lint := exec.Command(filepath.Join(dir, "shared-gcl"), args...)
 lint.Stdin, lint.Stdout, lint.Stderr = os.Stdin, os.Stdout, os.Stderr
 if err := lint.Run(); err != nil { return fmt.Errorf("lint: %w", err) }
 return nil
}
