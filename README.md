# go-linters

A shared Go lint command and library for Go repositories. The initial baseline
runs `go vet` across the target module, using its Go toolchain and module
configuration. Custom checks can be registered centrally in `lint.DefaultChecks`.
No organization-specific rules are enabled yet.

## Use in another repository

Install a specific published version (replace `vX.Y.Z` with a released tag):

```sh
go install github.com/rdpetrusek/go-linters/cmd/go-linters@vX.Y.Z
go-linters ./...
```

Run the command from the target module's root. With no arguments it checks
`./...`; arguments are Go package patterns. Findings and tool failures result in
a nonzero exit status. Pin a release in CI so updates to the shared policy are
intentional. This module requires Go 1.22 or newer. A release tag must be
published before the versioned install command will work.

For a GitHub Actions workflow in a consuming repository:

```yaml
name: Go lint
on: [push, pull_request]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - run: go install github.com/rdpetrusek/go-linters/cmd/go-linters@vX.Y.Z
      - run: go-linters ./...
```

## Add a shared check

Implement `lint.Check.Run`, accepting package patterns and writing diagnostics to
the supplied writer. Return an error for findings or execution failures. Add the
check to `DefaultChecks`; the runner executes all checks even when one fails.
Checks run in the consuming repository's working directory. Release a new version
and update consuming repositories' pins to roll out the change. For checks that
need type information, use a Go analysis framework and pin its dependencies in
this module. A consumer can also import `lint` and construct its own check list.

Run `go test ./...` and `go vet ./...` before publishing changes.
