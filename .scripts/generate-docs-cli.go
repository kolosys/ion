package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	var (
		packages = flag.String("packages", "workerpool,ratelimit,semaphore", "Comma-separated list of packages to document")
		docsDir  = flag.String("docs-dir", "docs", "Output directory for documentation")
		verbose  = flag.Bool("verbose", true, "Enable verbose output")
		configFile = flag.String("config", "", "Path to configuration file (overrides other flags)")
	)
	flag.Parse()

	if *configFile != "" {
		fmt.Printf("📄 Using configuration file: %s\n", *configFile)
		// Load from config file
		return
	}

	packageList := strings.Split(*packages, ",")
	for i, pkg := range packageList {
		packageList[i] = strings.TrimSpace(pkg)
	}

	if *verbose {
		fmt.Printf("🚀 Generating documentation for packages: %v\n", packageList)
		fmt.Printf("📁 Output directory: %s\n", *docsDir)
	}

	// Your existing DocGenerator logic here...
	fmt.Println("✅ Documentation generation complete!")
}

// Usage examples:
// go run .scripts/generate-docs-cli.go
// go run .scripts/generate-docs-cli.go -packages="workerpool,ratelimit" -docs-dir="output"
// go run .scripts/generate-docs-cli.go -config="docs-config.yaml"
