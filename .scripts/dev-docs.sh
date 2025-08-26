#!/bin/bash

# Quick development script for updating docs locally
# Usage: ./scripts/dev-docs.sh [package_name]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# If package name provided, only regenerate that package
if [ $# -eq 1 ]; then
    package_name="$1"
    echo "🔄 Updating documentation for $package_name..."
    
    # Check if package directory exists
    if [ ! -d "$ROOT_DIR/$package_name" ]; then
        echo "❌ Package directory not found: $package_name"
        echo "Available packages: workerpool, ratelimit, semaphore"
        exit 1
    fi
    
    # Regenerate docs for specific package
    cd "$ROOT_DIR"
    echo "📝 Regenerating API docs for $package_name..."
    
    # Create a temporary script to generate docs for one package
    cat > /tmp/generate-single-package.go << 'EOF'
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run generate-single-package.go <package_name>")
		os.Exit(1)
	}
	
	packageName := os.Args[1]
	
	// Use the existing DocGenerator but for single package
	// This is a simplified version - you could import the full generator
	fmt.Printf("Generating documentation for %s...\n", packageName)
	
	// For now, just run the full generator
	// In a real implementation, you'd modify the generator to accept package args
	fmt.Printf("✅ Generated documentation for %s\n", packageName)
}
EOF
    
    # For now, just run the full generator since it's fast enough
    go run scripts/generate-docs.go
    
    echo "✅ Documentation updated for $package_name"
    echo "📁 Updated: docs/$package_name/"
    
else
    # Full build
    echo "🚀 Running full documentation build..."
    ./scripts/build-docs.sh
fi

echo ""
echo "💡 Tips:"
echo "   - Edit templates in docs-templates/"
echo "   - Regenerate specific package: ./scripts/dev-docs.sh workerpool"
echo "   - Full rebuild: ./scripts/dev-docs.sh"
echo "   - View generated docs: open docs/ in your editor"
