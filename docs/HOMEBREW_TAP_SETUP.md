# Homebrew Tap Setup

This document describes how to set up a Homebrew tap for mcp-ssh-wingman.

## Overview

GoReleaser will automatically generate and publish the Homebrew formula to your tap repository when a new release is created.

## Creating the Tap Repository

1. Create a new GitHub repository named `homebrew-tap` under your account:
   ```bash
   # On GitHub, create a new repository: conallob/homebrew-tap
   ```

2. Clone the repository:
   ```bash
   git clone https://github.com/conallob/homebrew-tap.git
   cd homebrew-tap
   ```

3. Create the Formula directory:
   ```bash
   mkdir -p Formula
   ```

4. Create a README.md:
   ```bash
   cat > README.md << 'EOF'
   # Homebrew Tap for MCP SSH Wingman

   This is a Homebrew tap for [mcp-ssh-wingman](https://github.com/conallob/mcp-ssh-wingman).

   ## Installation

   ```bash
   brew tap conallob/tap
   brew install mcp-ssh-wingman
   ```

   ## Available Formulae

   - `mcp-ssh-wingman` - MCP Server for read-only access to Unix shell prompts via tmux
   EOF
   ```

5. Commit and push:
   ```bash
   git add .
   git commit -m "Initial tap setup"
   git push origin main
   ```

## GitHub Token Setup

For GoReleaser to publish to the tap repository, you need to create a GitHub token:

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Give it a name like "GoReleaser Homebrew Tap"
4. Select scopes:
   - `repo` (full control of private repositories)
   - `write:packages`
5. Generate the token and copy it
6. Add it to your mcp-ssh-wingman repository secrets:
   - Go to your mcp-ssh-wingman repository
   - Settings → Secrets and variables → Actions
   - Click "New repository secret"
   - Name: `HOMEBREW_TAP_GITHUB_TOKEN`
   - Value: paste your token
   - Click "Add secret"

## How It Works

When you push a new tag (e.g., `v1.0.0`), the GitHub Actions workflow will:

1. Trigger the release workflow
2. Run GoReleaser which will:
   - Build binaries for all target platforms
   - Create a GitHub release with the binaries
   - Generate a Homebrew formula
   - Commit the formula to your homebrew-tap repository

## Testing the Formula Locally

Before releasing, you can test the formula locally:

```bash
# Install from your tap
brew tap conallob/tap
brew install mcp-ssh-wingman

# Test the installation
mcp-ssh-wingman --version

# Uninstall
brew uninstall mcp-ssh-wingman
brew untap conallob/tap
```

## Formula Updates

The formula will be automatically updated by GoReleaser with each new release. You don't need to manually maintain it.

The generated formula will include:
- Binary installation
- tmux dependency
- Version information
- Checksums for verification
- Basic test command

## Manual Formula Template (Reference Only)

If you need to create a formula manually, here's a template (GoReleaser will handle this):

```ruby
class McpSshWingman < Formula
  desc "MCP Server for read-only access to Unix shell prompts via tmux"
  homepage "https://github.com/conallob/mcp-ssh-wingman"
  version "1.0.0"
  license "BSD-3-Clause"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/conallob/mcp-ssh-wingman/releases/download/v1.0.0/mcp-ssh-wingman_1.0.0_Darwin_arm64.tar.gz"
      sha256 "CHECKSUM_HERE"
    end
    if Hardware::CPU.intel?
      url "https://github.com/conallob/mcp-ssh-wingman/releases/download/v1.0.0/mcp-ssh-wingman_1.0.0_Darwin_x86_64.tar.gz"
      sha256 "CHECKSUM_HERE"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/conallob/mcp-ssh-wingman/releases/download/v1.0.0/mcp-ssh-wingman_1.0.0_Linux_arm64.tar.gz"
      sha256 "CHECKSUM_HERE"
    end
    if Hardware::CPU.intel?
      url "https://github.com/conallob/mcp-ssh-wingman/releases/download/v1.0.0/mcp-ssh-wingman_1.0.0_Linux_x86_64.tar.gz"
      sha256 "CHECKSUM_HERE"
    end
  end

  depends_on "tmux"

  def install
    bin.install "mcp-ssh-wingman"
  end

  test do
    system "#{bin}/mcp-ssh-wingman", "--version"
  end
end
```
