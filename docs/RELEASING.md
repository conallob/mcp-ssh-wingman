# Release Process

This document describes how to create a new release of mcp-ssh-wingman.

## Prerequisites

1. Ensure you have committed all changes
2. Ensure tests pass: `make test`
3. Ensure code is properly formatted: `make lint`
4. Set up Homebrew tap repository (see [HOMEBREW_TAP_SETUP.md](./HOMEBREW_TAP_SETUP.md))
5. Add `HOMEBREW_TAP_GITHUB_TOKEN` secret to repository

## Creating a Release

Releases are automated via GitHub Actions and GoReleaser. To create a new release:

### 1. Choose a version number

Follow [Semantic Versioning](https://semver.org/):
- **MAJOR** version for incompatible API changes
- **MINOR** version for new functionality in a backwards compatible manner
- **PATCH** version for backwards compatible bug fixes

Example: `v1.2.3`

### 2. Create and push a tag

```bash
# Ensure you're on main branch and up to date
git checkout main
git pull

# Create a tag (replace with your version)
git tag -a v1.0.0 -m "Release v1.0.0"

# Push the tag
git push origin v1.0.0
```

### 3. Wait for automation

The GitHub Actions workflow will automatically:

1. ✅ Build binaries for:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - FreeBSD (amd64, arm64)

2. ✅ Create a GitHub release with:
   - Compiled binaries
   - Checksums
   - Changelog

3. ✅ Update Homebrew tap:
   - Generate Homebrew formula
   - Commit to `homebrew-tap` repository

### 4. Verify the release

1. Check the [GitHub Actions run](https://github.com/conallob/mcp-ssh-wingman/actions)
2. Verify the [GitHub release](https://github.com/conallob/mcp-ssh-wingman/releases)
3. Check the [Homebrew tap](https://github.com/conallob/homebrew-tap)

### 5. Test installation

```bash
# Test Homebrew installation
brew update
brew upgrade mcp-ssh-wingman  # or brew install if not installed

# Verify version
mcp-ssh-wingman --version
```

## What Gets Built

### Platforms
- **Linux**: amd64, arm64
- **macOS**: amd64, arm64
- **FreeBSD**: amd64, arm64

### Archives
Each platform gets a `.tar.gz` archive containing:
- `mcp-ssh-wingman` binary
- `LICENSE`
- `README.md`
- `CLAUDE.md`

### Homebrew Formula
Automatically generated and published to `homebrew-tap` with:
- Multi-platform support (macOS/Linux, amd64/arm64)
- tmux dependency
- Installation instructions
- Version checksums

## Troubleshooting

### Release workflow fails

Check the GitHub Actions logs for errors. Common issues:
- Missing `HOMEBREW_TAP_GITHUB_TOKEN` secret
- Build errors (fix and push new tag)
- Network issues (re-run workflow)

### Homebrew formula not updated

1. Check that `HOMEBREW_TAP_GITHUB_TOKEN` is set correctly
2. Verify the token has `repo` permissions
3. Check that `homebrew-tap` repository exists

### Version not showing correctly

Ensure the tag is in the format `vX.Y.Z` (with the `v` prefix).

## Manual Release (Emergency)

If automation fails, you can create a release manually:

```bash
# Install GoReleaser
brew install goreleaser

# Create a release (dry run first)
goreleaser release --snapshot --clean

# Create actual release
export GITHUB_TOKEN="your-github-token"
export HOMEBREW_TAP_GITHUB_TOKEN="your-homebrew-token"
goreleaser release --clean
```

## Version Information

The version information is embedded at build time:
- **version**: Git tag (e.g., `v1.0.0`)
- **commit**: Git commit hash
- **date**: Build timestamp

View with: `mcp-ssh-wingman --version`
