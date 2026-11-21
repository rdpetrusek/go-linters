# go-linters

Reusable Go `go/analysis` analyzers and a module plugin for
[golangci-lint](https://golangci-lint.run/). The current example analyzer,
`todoowner`, reports `// TODO:` and `// TODO():` comments without an owner.
`multiimpl` reports named interfaces with fewer than two concrete implementations
in non-test packages of the same Go module. A named type and its pointer count as
one implementation. It scans the module using the default Go build environment;
test-only types, external modules, aliases, generic types and generic interfaces
are excluded. Constraint interfaces with type sets are skipped. Implementations
in other modules cannot be discovered. If any module package fails to load,
analysis fails rather than reporting incomplete counts.

## Use with golangci-lint

Install golangci-lint v2.7.2 (requires Go 1.24 or newer):

```sh
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.7.2
```

Ensure Go's binary directory (`GOBIN`, or `GOPATH/bin` when `GOBIN` is
unset) is on your `PATH`. Git and network access are needed to build the custom
binary.

In the consuming repository, create `.custom-gcl.yml`:

```yaml
version: v2.7.2
plugins:
  - module: github.com/rdpetrusek/go-linters
    import: github.com/rdpetrusek/go-linters/plugin
    version: vX.Y.Z
```

Replace `vX.Y.Z` with a published version of this module. A release containing
the plugin must be published before a release tag can be used. Build the
custom binary once from the consuming repository:

```sh
golangci-lint custom
```

This produces `./custom-gcl`. Enable the `shared` plugin in the consuming
repository's `.golangci.yml` alongside any standard linters it uses:

```yaml
version: "2"
linters:
  enable:
    - govet
    - shared
  settings:
    custom:
      shared:
        type: module
```

Merge these entries into an existing configuration rather than duplicating its
`linters` section. Then run:

```sh
./custom-gcl run ./...
```

The custom binary supports the normal golangci-lint CLI and config discovery;
use `--config` to select a file explicitly. Build again when the pinned
plugin version changes. Use the same pinned configuration and commands locally
and in CI. The `shared` plugin currently enables its registered analyzers as
a group; individual analyzers cannot be selected separately through it.

## Use with other Go analysis tools

Go tools that support `golang.org/x/tools/go/analysis` can import an analyzer
directly. For example:

```go
import (
    "github.com/rdpetrusek/go-linters/analyzers/todoowner"
    "golang.org/x/tools/go/analysis/multichecker"
)

func main() {
    multichecker.Main(todoowner.Analyzer)
}
```

Add a pinned version of this module to the tool's Go dependencies.

## Develop analyzers

Add reusable analyzer packages under `analyzers/` and register them in the
plugin's `BuildAnalyzers`. Run `go test ./...`; CI also exercises the
module-plugin build and a direct import from a separate Go module.
