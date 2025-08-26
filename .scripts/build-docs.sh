#!/bin/bash

# Main documentation build script
# Usage: ./scripts/build-docs.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "🚀 Building Ion documentation..."
echo "Working directory: $ROOT_DIR"

# Step 1: Generate API documentation from Go source
echo ""
echo "Step 1: Generating API documentation from Go source..."
cd "$ROOT_DIR"
go run scripts/generate-docs.go

# Step 2: Copy static template files  
echo ""
echo "Step 2: Copying documentation templates..."
bash scripts/copy-templates.sh

# Step 3: Generate examples documentation
echo ""
echo "Step 3: Generating examples documentation..."
bash scripts/generate-examples-docs.sh

# Step 4: Generate combined API reference
echo ""
echo "Step 4: Generating combined API reference..."
bash scripts/generate-api-reference.sh

echo ""
echo "✅ Documentation build complete!"
echo ""
echo "📁 Generated files:"
echo "   docs/getting-started.md"
echo "   docs/installation.md" 
echo "   docs/workerpool/"
echo "   docs/ratelimit/"
echo "   docs/semaphore/"
echo "   docs/reference/examples.md"
echo "   docs/reference/api.md"
echo ""
echo "🔗 GitBook will automatically sync these files when you push to GitHub"
