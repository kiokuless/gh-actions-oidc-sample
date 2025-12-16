# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
# Build server
go build -o bin/server ./cmd/server

# Build CLI
go build -o bin/cli ./cmd/cli

# Run server
./bin/server -audience "https://your-server.example.com" -allowed-owners "your-github-org"

# Run CLI
./bin/cli health -server http://localhost:8080
./bin/cli auth -server http://localhost:8080 -token $OIDC_TOKEN
```

## Architecture

This is a GitHub Actions OIDC authentication sample with two entry points:

- **cmd/server/**: HTTP server that validates OIDC tokens from GitHub Actions
  - Endpoints: `/health` (GET), `/auth` (POST)
  - Uses `pkg/oidc` for token verification

- **cmd/cli/**: CLI client with `auth` and `health` subcommands

- **pkg/oidc/**: Core OIDC verification logic
  - `verifier.go`: Token verification using `go-oidc/v3` against GitHub's OIDC issuer (`https://token.actions.githubusercontent.com`)
  - `policy.go`: Policy-based access control (allowed repos, owners, refs, actors, workflows)

## Key Concepts

The server verifies GitHub Actions OIDC tokens by:
1. Validating the JWT signature against GitHub's JWKS
2. Checking the audience claim matches configuration
3. Applying policy rules (AllowedRepositories, AllowedRepositoryOwners, etc.)
