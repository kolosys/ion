#!/bin/bash

# Generate combined API reference from individual package docs
# Usage: ./scripts/generate-api-reference.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
DOCS_DIR="$ROOT_DIR/docs"

echo "📚 Generating combined API reference..."

# Create reference directory
mkdir -p "$DOCS_DIR/reference"

# Start the API reference
cat > "$DOCS_DIR/reference/api.md" << 'EOF'
# Ion API Reference

Complete API documentation for all Ion packages.

## Overview

Ion provides the following packages:

- **[workerpool](../workerpool/)** - Bounded worker pool with context-aware submission and graceful shutdown
- **[ratelimit](../ratelimit/)** - Token bucket and leaky bucket rate limiters with configurable options  
- **[semaphore](../semaphore/)** - Weighted semaphore with configurable fairness modes

## Quick Navigation

EOF

# Add navigation links for each package
packages=("workerpool" "ratelimit" "semaphore")
for pkg in "${packages[@]}"; do
    if [ -f "$DOCS_DIR/$pkg/api-reference.md" ]; then
        echo "### $pkg" >> "$DOCS_DIR/reference/api.md"
        echo "" >> "$DOCS_DIR/reference/api.md"
        echo "- [Full $pkg API Reference](../$pkg/api-reference.md)" >> "$DOCS_DIR/reference/api.md"
        echo "- [Examples](../$pkg/examples.md)" >> "$DOCS_DIR/reference/api.md"
        echo "- [Overview](../$pkg/README.md)" >> "$DOCS_DIR/reference/api.md"
        echo "" >> "$DOCS_DIR/reference/api.md"
    fi
done

# Add summary sections for each package
echo "" >> "$DOCS_DIR/reference/api.md"
echo "## Package Summaries" >> "$DOCS_DIR/reference/api.md"
echo "" >> "$DOCS_DIR/reference/api.md"

for pkg in "${packages[@]}"; do
    if [ -f "$DOCS_DIR/$pkg/api-reference.md" ]; then
        echo "📝 Adding summary for $pkg"
        {
            echo "### $pkg"
            echo ""
            echo "[Full Documentation](../$pkg/api-reference.md)"
            echo ""
            
            # Extract functions and types sections (first 100 lines to avoid too much content)
            if head -100 "$DOCS_DIR/$pkg/api-reference.md" | grep -q "## Functions"; then
                echo "#### Key Functions"
                echo ""
                # Get function names from the api reference
                head -100 "$DOCS_DIR/$pkg/api-reference.md" | sed -n '/## Functions/,/## Types/p' | grep "^### " | head -5 | while read -r line; do
                    func_name=$(echo "$line" | sed 's/^### //')
                    echo "- **$func_name**"
                done
                echo ""
            fi
            
            if head -100 "$DOCS_DIR/$pkg/api-reference.md" | grep -q "## Types"; then
                echo "#### Key Types"
                echo ""
                # Get type names from the api reference  
                head -100 "$DOCS_DIR/$pkg/api-reference.md" | sed -n '/## Types/,/## Constants/p' | grep "^### " | head -5 | while read -r line; do
                    type_name=$(echo "$line" | sed 's/^### //')
                    echo "- **$type_name**"
                done
                echo ""
            fi
            
        } >> "$DOCS_DIR/reference/api.md"
    else
        echo "⚠️  API reference not found for $pkg"
    fi
done

# Add pkg.go.dev link
{
    echo "## External References"
    echo ""
    echo "- [pkg.go.dev Documentation](https://pkg.go.dev/github.com/kolosys/ion)"
    echo "- [GitHub Repository](https://github.com/kolosys/ion)"
    echo "- [Examples Directory](https://github.com/kolosys/ion/tree/main/examples)"
    echo ""
} >> "$DOCS_DIR/reference/api.md"

echo "✅ Combined API reference generated successfully"
