#!/bin/bash
set -e

# Script to help set up the Homebrew tap repository
# This creates the initial structure for homebrew-tap repository

GITHUB_USER="${1:-conallob}"
TAP_REPO="homebrew-tap"

echo "Setting up Homebrew tap for user: $GITHUB_USER"
echo ""

# Check if gh CLI is available
if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI (gh) is not installed."
    echo "Install it with: brew install gh"
    echo "Then authenticate with: gh auth login"
    exit 1
fi

# Check if authenticated
if ! gh auth status &> /dev/null; then
    echo "Error: Not authenticated with GitHub CLI"
    echo "Run: gh auth login"
    exit 1
fi

echo "Step 1: Creating GitHub repository..."
if gh repo create "$GITHUB_USER/$TAP_REPO" --public --description "Homebrew tap for MCP SSH Wingman" --clone=false; then
    echo "✅ Repository created: $GITHUB_USER/$TAP_REPO"
else
    echo "⚠️  Repository might already exist, continuing..."
fi

echo ""
echo "Step 2: Cloning repository..."
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

if git clone "https://github.com/$GITHUB_USER/$TAP_REPO.git"; then
    cd "$TAP_REPO"
else
    echo "Error: Failed to clone repository"
    exit 1
fi

echo ""
echo "Step 3: Creating directory structure..."
mkdir -p Formula

echo ""
echo "Step 4: Creating README..."
cat > README.md << EOF
# Homebrew Tap for MCP SSH Wingman

This is a Homebrew tap for [mcp-ssh-wingman](https://github.com/$GITHUB_USER/mcp-ssh-wingman).

## Installation

\`\`\`bash
brew tap $GITHUB_USER/tap
brew install mcp-ssh-wingman
\`\`\`

## Available Formulae

- \`mcp-ssh-wingman\` - MCP Server for read-only access to Unix shell prompts via tmux

## Automated Updates

The formulae in this tap are automatically updated by [GoReleaser](https://goreleaser.com/)
when new releases are published to the main repository.
EOF

echo ""
echo "Step 5: Creating .gitignore..."
cat > .gitignore << EOF
# macOS
.DS_Store

# Editor files
*.swp
*.swo
*~
.vscode/
.idea/
EOF

echo ""
echo "Step 6: Committing and pushing..."
git add .
git commit -m "Initial tap setup"
git push origin main

echo ""
echo "✅ Homebrew tap setup complete!"
echo ""
echo "Repository: https://github.com/$GITHUB_USER/$TAP_REPO"
echo ""
echo "Next steps:"
echo "1. Create a GitHub Personal Access Token:"
echo "   - Go to: https://github.com/settings/tokens/new"
echo "   - Note: 'GoReleaser Homebrew Tap'"
echo "   - Scopes: repo, write:packages"
echo ""
echo "2. Add the token to mcp-ssh-wingman repository secrets:"
echo "   - Go to: https://github.com/$GITHUB_USER/mcp-ssh-wingman/settings/secrets/actions"
echo "   - Click 'New repository secret'"
echo "   - Name: HOMEBREW_TAP_GITHUB_TOKEN"
echo "   - Value: [paste your token]"
echo ""
echo "3. Create a release to test:"
echo "   cd /path/to/mcp-ssh-wingman"
echo "   git tag -a v1.0.0 -m 'Release v1.0.0'"
echo "   git push origin v1.0.0"
echo ""

# Cleanup
cd /
rm -rf "$TEMP_DIR"
