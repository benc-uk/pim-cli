# PIM for Groups CLI

A command-line tool for managing Privileged Identity Management (PIM) for Groups in Microsoft Entra ID (Azure AD).

## Overview

`pim-cli` simplifies working with [PIM for Groups](https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management/concept-pim-for-groups) through the Microsoft Graph API. It allows you to:

> NOTE: PIM for Groups used to be known as PAG (Privileged Access Groups), but Microsoft loves changing names of things!

- List eligible PIM group memberships
- View currently active PIM group assignments
- Request activation of eligible group memberships

This is useful for users who need to frequently activate just-in-time access to privileged groups without navigating through the Azure Portal.

## Prerequisites

- Go 1.25 or later
- Azure credentials configured (see [Authentication](#authentication))
- Eligible PIM group assignments in your Entra ID tenant

## Installation

### Download from GitHub Releases

Pre-compiled binaries are available on the [GitHub Releases](https://github.com/benc-uk/pimg-cli/releases).
Select the appropriate binary for your operating system and architecture, download it, and place it in your system's PATH.

I could provide a script to automate this, but honestly, using `curl` or `wget` to grab the latest release from GitHub is straightforward enough.

### Build from Source

```bash
# Clone the repository
git clone https://github.com/benc-uk/pimg-cli.git
cd pimg-cli

# Install dependencies
make install

# Build the binary
make build
```

The compiled binary will be placed in `bin/pim-cli`.

### Cross-Platform Builds

```bash
make build-win   # Windows (amd64)
make build-mac   # macOS (amd64)
```

## Authentication

The CLI uses the Azure SDK's `DefaultAzureCredential` which attempts authentication in the following order:

1. **Azure CLI** - Uses credentials from `az login`
1. **Environment Variables** - Service principal credentials via `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_TENANT_ID`
1. **Workload Identity** - For containerized environments
1. **Managed Identity** - For Azure-hosted resources
1. **Azure Developer CLI** - Uses credentials from `azd auth login`

For local development, the easiest method is to authenticate via the Azure CLI:

```bash
az login
```

## Usage

### List Eligible PIM Groups

Show all PIM groups you're eligible to activate:

```bash
./bin/pim-cli list
```

List all Entra ID groups (not just PIM eligible):

```bash
./bin/pim-cli list --all
```

### View Active Assignments

Show your currently active PIM group assignments:

```bash
./bin/pim-cli active
```

### Request Activation

Activate an eligible PIM group membership:

```bash
./bin/pim-cli request --name "Group Name"
```

#### Request Options

| Flag         | Short | Description                                  | Default                |
| ------------ | ----- | -------------------------------------------- | ---------------------- |
| `--name`     | `-n`  | Name of the PIM group to activate (required) | -                      |
| `--reason`   | `-r`  | Justification for the activation request     | Auto-generated message |
| `--duration` | `-d`  | Duration of the activation                   | `12h`                  |

#### Examples

```bash
# Activate for 2 hours with a custom reason
./bin/pim-cli request -n "Production-Admins" -r "Incident response" -d 2h

# Activate for 30 minutes
./bin/pim-cli request -n "Database-Writers" -d 30m
```

## Development

### Make Targets

```bash
make help       # Show all available targets
make install    # Download dependencies including dev tools
make build      # Build for Linux
make lint       # Run golangci-lint
make tidy       # Tidy Go modules
make clean      # Remove build artifacts
make ver        # Show current version
```

### Project Structure

```
├── cmd/           # Cobra CLI commands
│   ├── root.go    # Root command and authentication
│   ├── list.go    # List eligible groups
│   ├── active.go  # Show active assignments
│   └── request.go # Request activation
├── pkg/
│   ├── graph/     # Microsoft Graph REST API client
│   └── pim/       # PIM-specific business logic
├── .dev/          # Development tools and configs
└── bin/           # Compiled binaries (git-ignored)
```

### Versioning

Version is derived from git tags at build time. If no tags exist, defaults to `0.0.0-dev`.

## License

See [LICENSE](LICENSE) for details.
