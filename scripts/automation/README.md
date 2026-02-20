# Automation Scripts

This directory contains automation scripts to improve development workflow for the Shinkansen Commerce project.

## Important Note for Nix Users

**Nix/NixOS systems require additional setup for file watching:**

The `proto-watch` feature requires `inotify-tools` to be available in your Nix profile.

### Installing inotify-tools on Nix

**Option 1: Add to your existing Nix configuration**

```nix
# In your configuration.nix or flake.nix
{pkgs, ...}:
  environment.systemPackages = with pkgs; [
    # ... other packages ...
    pkgs.inotify-tools
  ];
}
```

Then rebuild:
```bash
nix-env -iA nixpkgs.inotify-tools
```

**Option 2: Temporary installation**

```bash
# Add inotify-tools to your profile
nix-env -iA nixpkgs.inotify-tools

# Then the watch mode should work
export PATH=$HOME/.nix-profile/bin:$PATH
```

### Recommended Workflow for Nix Users

Due to Nix environment complexity, the **manual workflow** is recommended:

```bash
# 1. Edit proto files
# 2. Regenerate code
make gen

# 3. Commit
git add proto/ services/ docs/
git commit

# The pre-commit hook will automatically run code generation if needed
```

This approach is:
- ✅ More reliable (no environment-specific issues)
- ✅ Faster (no continuous background monitoring)
- ✅ Simpler to debug
- ✅ Works consistently across all systems

## Scripts

### pre-commit.sh
**Purpose**: Git pre-commit hook that automatically generates code when `.proto` files are committed.

**What it does**:
- Detects if any `.proto` files are staged for commit
- Runs `make proto-gen` to generate Go gRPC code
- Runs `make proto-gen-rust` to generate Rust protobuf code
- Runs `make proto-openapi-gen` to generate OpenAPI/Swagger docs
- Runs `make docs-gen-api` to generate API documentation
- Runs `make sqlc-gen` to generate SQL code
- Automatically stages generated files for commit

**Installation**:
```bash
make install-git-hooks
```

**Uninstallation**:
```bash
rm .git/hooks/pre-commit
```

### watch-proto.sh
**Purpose**: Watch `.proto` files for changes and automatically regenerate code.

**What it does**:
- Watches all files in `proto/` directory
- Automatically regenerates all code (protobuf, OpenAPI, docs) when files change
- Ignores temporary files (`*.swp`, `*.swo`, `*.bak`)

**Requirements**:
- **Linux**: `inotifywait` (usually available or via `apt install inotify-tools`)
- **Nix/NixOS**: `inotify-tools` (see installation instructions above)
- **macOS**: fswatch (via `brew install fswatch`)

**Usage**:
```bash
# Start watching
make proto-watch

# Or directly
bash scripts/automation/watch-proto.sh
```

**Note**: This is **experimental** on Nix systems. If it doesn't work, use the manual workflow instead.

### generate-api-docs.sh
**Purpose**: Generate VitePress API documentation from `.proto` files.

**What it does**:
- Parses proto service files to extract RPC methods and message types
- Generates markdown files in `docs/api/` for each service
- Creates an index with all services and their RPC method counts
- Includes implementation details and testing examples

**Usage**:
```bash
# Generate API documentation
make docs-gen-api

# Or directly
bash scripts/automation/generate-api-docs.sh
```

## Development Workflow

### Manual Workflow (Recommended for Nix)

```bash
1. Edit .proto files
2. Run: make gen
3. Update service code
4. Run tests
5. Commit
```

### Automated Workflow (Linux/macOS)

```bash
1. Run: make proto-watch        # Start watcher in background
2. Edit .proto files            # Code auto-generates
3. Update service code
4. Run tests
5. git add . && git commit   # Pre-commit hook ensures generated code is synced
```

### CI/CD Integration

GitHub Actions automatically:
1. Runs code generation on every push/PR that modifies proto files
2. Lints proto files with `buf`
3. Checks if generated code is up to date
4. Fails to build if generated code needs to be committed

## Benefits

### 1. Never Forget to Regenerate
- Pre-commit hook ensures generated code is always in sync
- CI/CD fails if proto and generated code drift

### 2. Immediate Feedback
- Watch mode provides instant code regeneration
- No manual command needed after proto edits

### 3. Consistent Documentation
- API docs generated from proto source of truth
- Documentation always matches actual implementation

### 4. Type Safety
- Proto-first approach ensures type safety across languages
- Go and Rust services share the same type definitions

## Troubleshooting

### Pre-commit Hook Issues

```bash
# Check if hook is installed and executable
ls -l .git/hooks/pre-commit

# Hook not running?
# Reinstall hook
make install-git-hooks

# Hook fails during commit?
# Check the output and see which step failed
# Usually it's a build error in make proto-gen or make proto-gen-rust
```

### Watch Mode Issues

```bash
# Watch mode not starting?
# Check if inotifywait is available
command -v inotifywait

# On Nix, install inotify-tools:
nix-env -iA nixpkgs.inotify-tools
export PATH=$HOME/.nix-profile/bin:$PATH

# On Linux (Debian/Ubuntu):
sudo apt install inotify-tools

# On macOS:
brew install fswatch
# Then update watch-proto.sh to use fswatch instead of inotifywatch

# Generated files not updating?
# Check if proto files are being watched in the right directory
# Make sure you're editing files in the proto/ directory
```

### Generated Code Issues

```bash
# Generated code not committing?
# The pre-commit hook should automatically add generated files
# If it fails, manually add:
make gen
git add gen/ services/*/internal/db/ services/gateway/docs/api/ docs/api/
git commit

# Proto changes not reflected in generated code?
# Manually regenerate:
make gen

# Then check the generated files to verify
ls gen/proto/go/
ls gen/proto/rust/src/
ls services/gateway/docs/api/
ls docs/api/
```

### Nix-Specific Issues

```bash
# Scripts not finding commands?
# Ensure your Nix profile is active:
source $HOME/.nix-profile/etc/profile.d/nix.sh

# Or add to your shell configuration:
# For bash: echo 'source $HOME/.nix-profile/etc/profile.d/nix.sh' >> ~/.bashrc
# For zsh: echo 'source $HOME/.nix-profile/etc/profile.d/nix.sh' >> ~/.zshrc

# Git hooks not working?
# The pre-commit hook uses #!/usr/bin/env bash for Nix compatibility
# This should work if your Nix profile is sourced

# Watch mode failing?
# Watch mode is experimental on Nix
# Use manual workflow instead (edit proto → make gen → commit)
```

## Maintenance

### Updating Scripts

After pulling changes from the repository, ensure git hooks are reinstalled:

```bash
make install-git-hooks
```

### Testing Automation

Test pre-commit hook:
```bash
# Create a dummy proto file
touch proto/test.proto

# Stage it
git add proto/test.proto

# Commit (hook should run)
git commit -m "test: check pre-commit hook"

# Clean up
git rm --cached proto/test.proto
```

Test API doc generation:
```bash
make docs-gen-api
ls docs/api/
```

## Best Practices

1. **Edit proto files first** - Always update `.proto` files before writing service code
2. **Run code generation** - Use `make gen` after proto changes
3. **Review generated code** - Verify types and methods match your expectations
4. **Write tests** - Use generated types in your tests
5. **Commit frequently** - Commit generated code together with proto changes
6. **Let automation help** - Use pre-commit hook and CI/CD to catch issues
7. **Check CI status** - Don't push if CI is failing
8. **Use manual workflow on Nix** - More reliable than watch mode

## Architecture Overview

The automation follows this flow:

```
┌─────────────────┐
│  Edit .proto   │
└──────┬──────────┘
       │
       ▼
┌─────────────────┐      ┌──────────────┐      ┌────────────────┐
│  Run make gen   │ ──► │  gen/proto/go/ │ ──► │  Docs/        │
└─────────────────┘      └──────────────┘      └────────────────┘
       │                       │                      │
       ▼                       ▼                      ▼
┌─────────────────┐      ┌──────────────┐      ┌────────────────┐
│  Services/      │      │  Pre-commit   │      │  GitHub CI     │
│  src/          │      │  Hook (if     │      │  (On push)     │
│                 │      │   .proto staged)│ └────────────────┘
└─────────────────┘      └──────────────┘
       │
       ▼
┌─────────────────┐
│  git push      │
└─────────────────┘
```
