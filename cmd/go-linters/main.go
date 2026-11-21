package main

import (
 _ "embed"
 "flag"
 "fmt"
 "io"
 "os"
 "os/exec"
 "path/filepath"
 "runtime"
 "runtime/debug"
 "strings"
)

const module = "github.com/rdpetrusek/go-linters"
const golangciVersion = "v2.7.2"

//go:embed config.yml
var policy []byte

func main() {
 destination := flag.String("dir", ".bin", "directory for the custom binary and shared configuration")
 flag.Parse()
 if flag.NArg() != 0 { fmt.Fprintln(os.Stderr, "unexpected arguments"); os.Exit(2) }
 if err := install(*destination); err != nil { fmt.Fprintln(os.Stderr, "go-linters:", err); os.Exit(1) }
}

func install(destination string) error {
 source := os.Getenv("GO_LINTERS_SOURCE")
 info, ok := debug.ReadBuildInfo()
 if !ok { return fmt.Errorf("cannot determine module version") }
 version := info.Main.Version
 if source == "" && (version == "" || version == "(devel)") { return fmt.Errorf("use a released version or set GO_LINTERS_SOURCE to the absolute source directory") }
 dir, err := os.MkdirTemp("", "go-linters-*")
 if err != nil { return err }
 defer func() { _ = os.RemoveAll(dir) }()
 bin := filepath.Join(dir, "bin")
 if err := os.Mkdir(bin, 0755); err != nil { return err }
 executable := "golangci-lint"
 customName := "golangci-lint"
 if runtime.GOOS == "windows" { executable += ".exe"; customName += ".exe" }
 install := exec.Command("go", "install", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@"+golangciVersion)
 install.Env = append(os.Environ(), "GOBIN="+bin)
 install.Stdout, install.Stderr = os.Stdout, os.Stderr
 if err := install.Run(); err != nil { return fmt.Errorf("install golangci-lint: %w", err) }
 plugin := "    version: " + version
 if source != "" {
  if !filepath.IsAbs(source) || strings.ContainsAny(source, "\r\n\"") { return fmt.Errorf("GO_LINTERS_SOURCE must be an absolute path without quotes or newlines") }
  plugin = "    path: \"" + source + "\""
 }
 build := "version: " + golangciVersion + "\nname: golangci-lint\ndestination: " + dir + "\nplugins:\n  - module: " + module + "\n    import: " + module + "/plugin\n" + plugin + "\n"
 if err := os.WriteFile(filepath.Join(dir, ".custom-gcl.yml"), []byte(build), 0600); err != nil { return err }
 custom := exec.Command(filepath.Join(bin, executable), "custom", "-v")
 custom.Dir = dir
 custom.Stdout, custom.Stderr = os.Stdout, os.Stderr
 if err := custom.Run(); err != nil { return fmt.Errorf("build custom golangci-lint: %w", err) }
 if err := os.MkdirAll(destination, 0755); err != nil { return err }
 input, err := os.Open(filepath.Join(dir, customName))
 if err != nil { return err }
 defer func() { _ = input.Close() }()
 output, err := os.CreateTemp(destination, ".go-linters-*")
 if err != nil { return err }
 tmp := output.Name()
 defer func() { _ = os.Remove(tmp) }()
 if _, err := io.Copy(output, input); err != nil { _ = output.Close(); return err }
 if err := output.Chmod(0755); err != nil { _ = output.Close(); return err }
 if err := output.Close(); err != nil { return err }
 if err := os.Rename(tmp, filepath.Join(destination, customName)); err != nil { return err }
 if err := os.WriteFile(filepath.Join(destination, "go-linters.yml"), policy, 0644); err != nil { return err }
 fmt.Printf("Installed %s and shared configuration in %s\n", customName, destination)
 return nil
}
