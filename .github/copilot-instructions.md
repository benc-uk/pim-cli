# pimg-cli Development Guide

## Project Overview

CLI tool for interacting with Microsoft Graph API, specifically working with PIM for Groups. Built with Go 1.25, uses Azure SDK for authentication and direct REST API calls for Graph access.

This URL explains PIM for Groups in Microsoft Graph: https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management/concept-pim-for-groups you should be familiar with the concepts there to understand what this tool does.

**Module:** `github.com/benc-uk/pimg-cli`

## Architecture

Standard Cobra CLI structure:

- **main.go**: Entry point, calls `cmd.Execute()` with version
- **cmd/**: Cobra commands
  - `root.go` - Root command, authentication helper, global flags
  - `list.go` - List PIM groups command
  - `active.go` - Show active assignments command
  - `request.go` - Request group activation command
- **pkg/graph/**: Microsoft Graph REST API client (direct HTTP calls)
- **pkg/pim/**: PIM-specific business logic
- **bin/pimg**: Compiled binary output

## Development Workflow

### Build & Run

```bash
make build          # Builds to bin/pimg with version from git tags
make install        # Downloads all dependencies including dev tools
make help           # Shows all available make targets
```

### Version Handling

Version is injected at build time via git tags using ldflags: `-X 'main.version=$(VERSION)'`

- Derives from `git describe --tags --abbrev=0 --dirty=-dev`
- Falls back to `0.0.0-dev` if no tags exist

### Development Tools

Tools are managed in a separate module at [.dev/tools.mod](.dev/tools.mod):

- `golangci-lint/v2` - Linting (config: [.dev/golangci.yaml](.dev/golangci.yaml))
- `goimports` - Import formatting

Run tools via: `go tool -modfile=.dev/tools.mod <tool-name>`

### Linting Configuration

Custom golangci-lint rules ([.dev/golangci.yaml](.dev/golangci.yaml)):

- Max line length: 150 characters
- Enabled: bodyclose, gosec, nilerr, nilnil, revive, staticcheck
- Ignores G404 (weak random), G402 (TLS MinVersion), var-naming issues

## Project Conventions

### File Organization

- Compiled binaries → `bin/`
- Dev tooling & configs → `.dev/`
- Main application code → `cmd/`
- Environment variables → `.env` (not tracked, use [.dev/.env.sample](.dev/.env.sample))

### Build Artifacts (Ignored)

Per [.gitignore](.gitignore): `bin/`, `res/`, `misc/`, `tools/`, `.vite/`, `web/main.wasm`, `.env`

## Key Dependencies

- **Azure SDK** (`azcore`, `azidentity`): Authentication foundation
- **Cobra**: CLI framework

## Environment Setup

Copy [.dev/.env.sample](.dev/.env.sample) to `.env` in project root and configure required values for Azure/Graph authentication.

## Notes

- Project includes commented-out makefile targets suggesting potential future features: testing, cross-platform builds, web/WASM target
- Uses Go 1.25 features (current edge version as of 2026)
