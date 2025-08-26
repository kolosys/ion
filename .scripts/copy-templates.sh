#!/bin/bash

# Copy documentation templates to docs directory
# Usage: ./scripts/copy-templates.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
DOCS_DIR="$ROOT_DIR/docs"
TEMPLATES_DIR="$ROOT_DIR/docs-templates"

echo "📄 Copying documentation templates..."

# Check if templates directory exists
if [ ! -d "$TEMPLATES_DIR" ]; then
    echo "❌ Templates directory not found: $TEMPLATES_DIR"
    exit 1
fi

# Create docs directory if it doesn't exist
mkdir -p "$DOCS_DIR"

# Copy each template file
copied_files=0
for template_file in "$TEMPLATES_DIR"/*.md; do
    if [ -f "$template_file" ]; then
        filename=$(basename "$template_file")
        echo "📝 Copying $filename"
        cp "$template_file" "$DOCS_DIR/"
        copied_files=$((copied_files + 1))
    fi
done

if [ $copied_files -eq 0 ]; then
    echo "⚠️  No template files found in $TEMPLATES_DIR"
else
    echo "✅ Copied $copied_files template files successfully"
fi
