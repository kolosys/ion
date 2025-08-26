#!/bin/bash

# Generate examples documentation from the examples directory
# Usage: ./scripts/generate-examples-docs.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
DOCS_DIR="$ROOT_DIR/docs"

echo "🔍 Generating examples documentation..."

# Create reference directory
mkdir -p "$DOCS_DIR/reference"

# Start the examples documentation
cat > "$DOCS_DIR/reference/examples.md" << 'EOF'
# Examples

Complete working examples for each Ion component.

EOF

# Check if examples directory exists
if [ ! -d "$ROOT_DIR/examples" ]; then
    echo "⚠️  No examples directory found"
    echo "No examples available yet." >> "$DOCS_DIR/reference/examples.md"
    exit 0
fi

# Process each example directory
found_examples=false
for example_dir in "$ROOT_DIR/examples"/*; do
    if [ -d "$example_dir" ]; then
        found_examples=true
        example_name=$(basename "$example_dir")
        echo "📝 Processing example: $example_name"
        
        {
            echo "## $example_name"
            echo ""
            
            # Check for README in example directory
            if [ -f "$example_dir/README.md" ]; then
                echo "### Description"
                echo ""
                # Skip the first line (title) and add the rest
                tail -n +2 "$example_dir/README.md"
                echo ""
            fi
            
            echo "### Code"
            echo ""
            
            if [ -f "$example_dir/main.go" ]; then
                echo '```go'
                cat "$example_dir/main.go"
                echo '```'
                echo ""
            else
                echo "No main.go file found in this example."
                echo ""
            fi
            
            echo "### Running this example"
            echo ""
            echo '```bash'
            echo "cd examples/$example_name"
            echo "go run main.go"
            echo '```'
            echo ""
            
            echo "[View on GitHub](https://github.com/kolosys/ion/tree/main/examples/$example_name)"
            echo ""
        } >> "$DOCS_DIR/reference/examples.md"
    fi
done

if [ "$found_examples" = false ]; then
    echo "No examples available yet." >> "$DOCS_DIR/reference/examples.md"
    echo "⚠️  No example directories found in examples/"
else
    echo "✅ Examples documentation generated successfully"
fi
