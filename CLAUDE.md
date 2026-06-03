# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build      # compile to ./build/routest (with version ldflags)
make test       # go test -v ./...
make testcov    # coverage report → coverage.out
make vet        # go vet ./...
make lint       # golangci-lint run -v --timeout=15m ./...
make install    # copy binary to ~/.local/bin
make clean      # remove build artifacts
```

## Architecture

`routest` is a single-binary CLI that tests Alertmanager routing configurations — given a set of labels, it reports which receiver(s) an alert would be routed to.

**Entry point:** `cmd/routest/main.go`
- Parses `-file`, `-labels`, and `-version` flags
- Reads Alertmanager YAML config from a file or stdin
- Unmarshals config using `gopkg.in/yaml.v3` into `github.com/prometheus/alertmanager/config.Config`
- Builds a route dispatch tree via `dispatch.NewRoute(c.Route, nil)`
- Calls `.Match(labelSet)` to find matching routes and prints their receivers

**Build info:** `internal/buildinfo/buildinfo.go` — `Version`, `Name`, and `GitSHA` are injected via ldflags at build time (see `Makefile` `LDFLAGS`).

The Alertmanager `dispatch` package handles route tree traversal including `continue` semantics; `Match()` returns all routes the alert should be delivered to.
