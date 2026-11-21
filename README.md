# go-linters

A custom [golangci-lint](https://golangci-lint.run/) binary with shared Go
analyzers compiled in. It accepts the usual golangci-lint commands and
configuration files. The supplied configuration enables `errcheck`, `govet`,
`staticcheck`, `unused`, and the custom `shared` plugin.

## Install

From the root of a consuming Go module, install a pinned release:

```sh
go run github.com/rdpetrusek/go-linters/cmd/go-linters@vX.Y.Z -dir .bin
```

Replace `vX.Y.Z` with a published release tag. A release containing the
installer must be published before the command works with a tag. Installation
requires Go 1.24 or newer, Git, and network access. It builds golangci-lint
v2.7.2 with the custom plugin and writes these files:

- `.bin/golangci-lint`: the custom executable.
- `.bin/go-linters.yml`: the shared configuration.

The default install directory is `.bin`; use `-dir` to choose another.
Installation replaces a `golangci-lint` executable already in that directory.
The build happens during installation, not on each lint run. Reinstall when
you change the pinned release. Add `.bin/` to the consuming repository's
`.gitignore` if binaries should remain local.

## Run

Use the installed configuration from the root of the consuming module:

```sh
.bin/golangci-lint run --config .bin/go-linters.yml ./...
```

Alternatively, pass your own configuration with `--config`, or use the usual
`.golangci.yml` discovery:

```sh
.bin/golangci-lint run ./...
```

A custom binary makes the analyzers available; the configuration must enable
them. To use the shared plugin in your own v2 configuration, merge the following
into its existing `linters` section:

```yaml
version: "2"
linters:
  enable:
    - shared
  settings:
    custom:
      shared:
        type: module
```

The current plugin groups custom analyzers under `shared`. It includes an
example `todoowner` analyzer that reports `// TODO:` and `// TODO():`
comments without an owner. All analyzers registered under `shared` run when
that plugin is enabled; they are not individually selectable through this
plugin configuration.

## CI

Use the same pinned version and command locally and in CI. For example, after
checking out a repository and setting up Go in GitHub Actions:

```yaml
- run: go run github.com/rdpetrusek/go-linters/cmd/go-linters@vX.Y.Z -dir .bin
- run: .bin/golangci-lint run --config .bin/go-linters.yml ./...
```

Replace the placeholder with a published version. The installer runs once per
setup; the binary can be reused for subsequent lint runs.

## Developing shared analyzers

Add Go analyzers to the plugin in `plugin/` and register them in
`BuildAnalyzers`. Repositories receive changes when they install a version
that contains them. In this repository, `make test` runs tests, `make install`
builds the local version into `.bin`, and `make lint` runs that binary.
