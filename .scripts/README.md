# Documentation Scripts

This directory contains reusable scripts for generating and managing Ion's documentation. These scripts work both locally and in CI/CD pipelines.

## Scripts Overview

### `build-docs.sh` - Main Build Script
Generates complete documentation from Go source code and templates.

```bash
./scripts/build-docs.sh
```

**What it does:**
1. Generates API docs from Go source (`generate-docs.go`)
2. Copies static templates (`copy-templates.sh`)
3. Generates examples documentation (`generate-examples-docs.sh`)
4. Creates combined API reference (`generate-api-reference.sh`)

### `dev-docs.sh` - Development Helper
Quick script for local development and testing.

```bash
# Full rebuild
./scripts/dev-docs.sh

# Update specific package only
./scripts/dev-docs.sh workerpool
```

### Individual Scripts

#### `copy-templates.sh`
Copies markdown files from `docs-templates/` to `docs/`.

```bash
./scripts/copy-templates.sh
```

#### `generate-examples-docs.sh`
Generates `docs/reference/examples.md` from the `examples/` directory.

```bash
./scripts/generate-examples-docs.sh
```

#### `generate-api-reference.sh`
Creates combined API reference at `docs/reference/api.md`.

```bash
./scripts/generate-api-reference.sh
```

#### `generate-docs.go`
Go program that parses source code and generates package-specific documentation.

```bash
go run scripts/generate-docs.go
```

## Usage Patterns

### Local Development
```bash
# Make changes to Go code or templates
vim workerpool/pool.go
vim docs-templates/getting-started.md

# Regenerate docs
./scripts/dev-docs.sh

# Check the results
ls docs/
```

### CI/CD Pipeline
The GitHub workflow uses the main build script:

```yaml
- name: Build documentation
  run: ./scripts/build-docs.sh
```

### Template Management
1. Edit files in `docs-templates/`
2. Run `./scripts/copy-templates.sh` to update `docs/`
3. GitBook syncs from `docs/`

## File Structure

```
scripts/
├── README.md                    # This file
├── build-docs.sh               # Main build script
├── dev-docs.sh                 # Development helper
├── copy-templates.sh           # Template copying
├── generate-examples-docs.sh   # Examples documentation
├── generate-api-reference.sh   # Combined API reference
└── generate-docs.go            # Go source parser

docs-templates/
├── getting-started.md          # Static template
└── installation.md             # Static template

docs/                           # Generated output (GitBook syncs this)
├── getting-started.md          # Copied from template
├── installation.md             # Copied from template
├── workerpool/                 # Generated from Go source
├── ratelimit/                  # Generated from Go source
├── semaphore/                  # Generated from Go source
└── reference/
    ├── examples.md             # Generated from examples/
    └── api.md                  # Combined API reference
```

## Benefits

✅ **Reusable** - Same scripts work locally and in CI  
✅ **Maintainable** - Easy to update and extend  
✅ **Fast** - Individual scripts for quick updates  
✅ **Consistent** - Same output everywhere  
✅ **Debuggable** - Run locally to test changes  

## Adding New Documentation

### New Static Page
1. Create `docs-templates/new-page.md`
2. Add to `copy-templates.sh` if needed
3. Update `docs/SUMMARY.md` for GitBook navigation

### New Generated Content
1. Add script to generate the content
2. Make it executable: `chmod +x scripts/new-script.sh`
3. Add to `build-docs.sh`
4. Test locally: `./scripts/new-script.sh`

## Troubleshooting

### Script Not Executable
```bash
chmod +x scripts/*.sh
```

### Missing Dependencies
Make sure Go is installed:
```bash
go version
```

### Windows/WSL Issues
Scripts are designed to work in Git Bash, WSL, or native Linux/macOS terminals.
