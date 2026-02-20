#!/usr/bin/env bash
# Watch proto files and auto-regenerate code
# This script watches for changes in proto/ directory
# and automatically regenerates all generated code (Go, Rust, OpenAPI, API docs, SQL)
#
# This is an EXPERIMENTAL feature on Nix-based systems
# For stable use, the manual workflow is recommended:
#   1. Edit proto files
#   2. Run: make gen
#   3. Commit (pre-commit hook will ensure generated code is synced)
#
# Usage:
#   make proto-watch
#
# Or directly:
#   bash scripts/automation/watch-proto.sh
#
# On Nix systems, you may need to install inotify-tools:
#   nix-env -iA nixpkgs.inotify-tools
#   or add to configuration.nix
#
# Then set up PATH:
#   export PATH=$HOME/.nix-profile/bin:$PATH
#

set -e

echo "👀 Proto File Watcher (Experimental)"
echo ""

# Try to find inotifywait in common Nix locations
INOTIFYWAIT=""
for dir in "$HOME/.nix-profile/bin" "/run/current-system/sw/bin" "/nix/var/nix/profiles/default/bin" "/etc/profiles/per-user/afasari/bin" "/usr/local/bin" "/usr/bin"; do
    if [ -x "$dir/inotifywait" ]; then
        INOTIFYWAIT="$dir/inotifywait"
        break
    fi
done

# Check if inotifywait is available
if [ -z "$INOTIFYWAIT" ]; then
    echo "⚠️  Warning: inotifywait not found in PATH or common locations"
    echo ""
    echo "Searched in:"
    echo "  - $HOME/.nix-profile/bin"
    echo "  - /run/current-system/sw/bin"
    echo "  - /nix/var/nix/profiles/default/bin"
    echo "  - /etc/profiles/per-user/afasari/bin"
    echo "  - /usr/local/bin"
    echo "  - /usr/bin"
    echo ""
    echo "Auto-regeneration requires inotifywait."
    echo ""
    echo "Solutions:"
    echo "  1. Install inotify-tools in your Nix configuration:"
    echo "     environment.systemPackages = [ pkgs.inotify-tools ];"
    echo "     Then run: nix-env -iA nixpkgs.inotify-tools"
    echo "  2. Or use manual workflow:"
    echo "     - Edit proto files"
    echo "     - Run 'make gen'"
    echo "     - Pre-commit hook will ensure code is synced"
    echo ""
    echo "⚠️  Switching to manual workflow mode..."
    
    # Manual mode: Just show reminder
    echo "💡 Manual workflow reminder:"
    echo "   Edit proto files, then run: make gen"
    exit 1
fi

echo "✅ inotifywait found at: $INOTIFYWAIT"
echo "📂 Watching: proto/"
echo "🔄 Auto-regenerate on file save"
echo "🛑 Press Ctrl+C to stop"
echo ""

# Function to handle file change events
on_file_changed() {
    echo ""
    echo "📝 Proto files changed - regenerating code..."
    
    # Run all generation targets
    if make proto-gen; then
        echo "   ✅ Go protobuf code generated"
    else
        echo "   ❌ Go protobuf generation failed"
    fi
    
    if make proto-gen-rust; then
        echo "   ✅ Rust protobuf code generated"
    else
        echo "   ❌ Rust protobuf generation failed"
    fi
    
    if make proto-openapi-gen; then
        echo "   ✅ OpenAPI docs generated"
    else
        echo "   ❌ OpenAPI generation failed"
    fi
    
    if make docs-gen-api; then
        echo "   ✅ API documentation generated"
    else
        echo "   ❌ API documentation generation failed"
    fi
    
    if make sqlc-gen; then
        echo "   ✅ SQL code generated"
    else
        echo "   ❌ SQL code generation failed"
    fi
    
    echo ""
    echo "✅ Code regeneration completed!"
    echo "   Watching for next change..."
    echo ""
}

# Set up inotifywait command
# Watch for create, modify, move, delete events on all files and directories in proto/
# --excludeq suppresses verbose output
# --format format the output as needed by our parsing
# --monitor recursively watch directories
# --event specify which events to monitor

echo "👀 Starting watcher..."
"$INOTIFYWAIT" \
    --excludeq \
    --format '%w %e %f' \
    --event create,modify,move,delete \
    --monitor \
    -r proto/ \
    2>/dev/null | while read -r line; do
    on_file_changed "$line"
done

echo ""
echo "👀 Watcher stopped"
