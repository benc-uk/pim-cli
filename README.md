# PIM for Groups CLI

![GitHub release](https://img.shields.io/github/v/release/benc-uk/pim-cli)
![CI Status](https://img.shields.io/github/actions/workflow/status/benc-uk/pim-cli/ci.yaml)
![GitHub last commit](https://img.shields.io/github/last-commit/benc-uk/pim-cli)
![License](https://img.shields.io/github/license/benc-uk/pim-cli)

## Overview

A command-line tool for those of us that are either lazy or busy, or more likely, both.
This CLI allows you to request activations for [Privileged Identity Management (PIM) group](https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management/concept-pim-for-groups) memberships, along with listing eligible and active assignments and some other helpful tasks.

> NOTE: PIM for Groups used to be known as PAG (Privileged Access Groups), but Microsoft loves changing names of things!

Current features include:

- List eligible PIM group memberships
- Request activation of an eligible PIM group
- View currently active PIM group assignments, including expiry times
- View pending activation requests

This is useful for users who need to frequently activate just-in-time access to privileged groups without navigating through the Azure Portal.

## Prerequisites

- Azure credentials configured (see [Authentication](#authentication))
- Eligible PIM group assignments in your Entra ID tenant
- Azure CLI installed and logged in
- Go 1.25 or later - for development and building from source only

## Installation

### Download from GitHub Releases

Pre-compiled binaries are available on the [GitHub Releases](https://github.com/benc-uk/pim-cli/releases).
Select the appropriate binary for your operating system and architecture, download it, and place it in your system's PATH.

I could provide a script to automate this, but honestly, using `curl` or `wget` to grab the latest release from GitHub is straightforward enough. I'm not your mum ok.

### Build from Source

```bash
# Clone the repository
git clone https://github.com/benc-uk/pim-cli.git
cd pim-cli

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

The CLI uses the Azure SDK's `DefaultAzureCredential` which attempts authentication via a range of methods, but 99% of the time you'll want to use the Azure CLI method. Simply ensure you're logged in with the Azure CLI (to the correct tenant if you have multiple) before running the tool.

I have not tested any other authentication scenarios, as I can't be bothered.

## Usage

### List Eligible PIM Groups

Show all PIM groups you're eligible to activate:

```bash
pim-cli list
```

### View Active Assignments

Show your currently active PIM group assignments:

```bash
pim-cli active
```

### View Pending Requests

Show your pending PIM group activation requests:

```bash
pim-cli pending
```

### View Status (Active + Pending)

Show both active and pending PIM group assignments in one command:

```bash
pim-cli status
```

### Global Options

| Flag      | Short | Description                                 |
| --------- | ----- | ------------------------------------------- |
| `--quiet` | `-q`  | Less verbose output for a more compact view |

### Request Activation

Activate an eligible PIM group membership:

```bash
pim-cli request --name "Group Name"
```

#### Request Options

| Flag         | Short | Description                                     | Default                |
| ------------ | ----- | ----------------------------------------------- | ---------------------- |
| `--name`     | `-n`  | Name of the PIM group to activate (required)    | -                      |
| `--reason`   | `-r`  | Justification for the activation request        | Auto-generated message |
| `--duration` | `-d`  | Duration of the activation                      | `12h`                  |
| `--role`     | `-o`  | Role name to activate (e.g., 'Member', 'Owner') | `Member`               |

#### Examples

```bash
# Activate for 2 hours with a custom reason
pim-cli request -n "Production-Admins" -r "Incident response" -d 2h

# Activate for 30 minutes
pim-cli request -n "Database-Writers" -d 30m

# Activate as Owner instead of Member
pim-cli request -n "Production-Admins" --role Owner
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
│   ├── pending.go # Show pending requests
│   ├── status.go  # Show active + pending
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
