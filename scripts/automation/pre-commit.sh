#!/usr/bin/env bash
# Pre-commit hook to auto-generate code from proto files
# This ensures generated code is always in sync with proto definitions

set -e

echo "🔄 Running pre-commit code generation..."

# Check if any proto files were changed
STAGED_PROTO_FILES=$(git diff --cached --name-only | grep -E '\.proto$' || true)

if [ -z "$STAGED_PROTO_FILES" ]; then
    echo "✅ No proto files staged. Skipping code generation."
    exit 0
fi

echo "📝 Proto files staged for commit:"
echo "$STAGED_PROTO_FILES"

# Run code generation
echo ""
echo "🔄 Generating Go protobuf code..."
make proto-gen

echo "🔄 Generating Rust protobuf code..."
make proto-gen-rust

echo "🔄 Generating OpenAPI docs..."
make proto-openapi-gen

echo "🔄 Generating API documentation from proto..."
make docs-gen-api

echo "🔄 Generating SQL code..."
make sqlc-gen

# Add generated files to staging
echo ""
echo "📦 Adding generated files to commit..."
git add gen/proto/go/
git add gen/proto/rust/
git add services/gateway/docs/api/
git add services/*/internal/db/

echo ""
echo "✅ Code generation completed successfully!"
echo "📝 Please review the generated changes before committing."
