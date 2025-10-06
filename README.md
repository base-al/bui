# Bui - Unified CLI for Base Stack

A modern, unified CLI tool for Base Stack development. Generate backend modules (Go) and frontend modules (Nuxt/TypeScript) with a single tool, powered by [Mamba](https://github.com/base-go/mamba).

[![Go Version](https://img.shields.io/github/go-mod/go-version/base-al/bui)](https://golang.org/dl/)
[![Release](https://img.shields.io/github/v/release/base-al/bui)](https://github.com/base-al/bui/releases)
[![License](https://img.shields.io/github/license/base-al/bui)](LICENSE)

## Features

- **Unified Interface**: Single CLI for both backend and frontend generation
- **Backend Generation**: Go modules with models, services, controllers, and validators
- **Frontend Generation**: Nuxt modules with TypeScript types, Pinia stores, Vue components, and pages
- **Modern UX**: Powered by Mamba for enhanced terminal features (colors, spinners, progress bars)
- **Smart Conventions**: Automatically infers field types and relationships

## Installation

### Option 1: Install Script (Recommended)

```bash
# Interactive installation (choose local or global)
curl -sSL https://raw.githubusercontent.com/base-al/bui/main/install.sh | bash
```

The script will ask you to choose:
- **Local install** (`~/.base/bin`) - No sudo required, add `~/.base/bin` to PATH
- **Global install** (`/usr/local/bin`) - Requires sudo, available system-wide

### Option 2: Go Install

```bash
# Install specific version
go install github.com/base-al/bui@v0.0.1

# Install latest version
go install github.com/base-al/bui@latest
```

### Option 3: Download Binary

Download pre-built binaries from [GitHub Releases](https://github.com/base-al/bui/releases) for your platform:
- macOS (darwin_amd64, darwin_arm64)
- Linux (linux_amd64, linux_arm64)
- Windows (windows_amd64, windows_arm64)

## Usage

### Generate Backend Module (Go)

```bash
# Generate a backend module with API endpoints
bui generate backend product name:string price:float stock:int description:text

# Short aliases
bui g be product name:string price:float
bui g api product name:string price:float
```

**Generates:**
- `app/models/product.go` - GORM model
- `app/products/service.go` - Business logic
- `app/products/controller.go` - HTTP handlers
- `app/products/module.go` - Module registration
- `app/products/validator.go` - Input validation

### Generate Frontend Module (Nuxt/TypeScript)

```bash
# Generate a frontend module with UI
bui generate frontend product name:string price:float stock:int description:text

# Short aliases
bui g fe product name:string price:float
bui g ui product name:string price:float
```

**Generates:**
- `admin/app/modules/products/types/product.ts` - TypeScript interfaces
- `admin/app/modules/products/stores/products.ts` - Pinia store with CRUD
- `admin/app/modules/products/components/ProductTable.vue` - Data table
- `admin/app/modules/products/components/ProductFormModal.vue` - Form modal
- `admin/app/modules/products/utils/formatters.ts` - Formatting utilities
- `admin/app/pages/app/products/index.vue` - List page
- `admin/app/pages/app/products/[id].vue` - Detail page

## Supported Field Types

### Basic Types
- `string` - Text field
- `text` - Textarea field
- `int`, `uint` - Integer numbers
- `float`, `float32`, `float64` - Decimal numbers
- `bool` - Boolean/checkbox

### Smart Field Detection
The CLI intelligently detects field purposes by name:
- `email` - Email input
- `password` - Password input
- `url`, `link` - URL input
- `description`, `content`, `bio` - Textarea
- `status`, `category`, `type` - Select dropdown
- `*_id` (ending with _id) - Foreign key (number)

## Other Commands

```bash
# Show version
bui version

# Start the application (backend)
bui start
```

## Why Mamba?

Bui uses [Mamba](https://github.com/base-go/mamba), a modern drop-in replacement for Cobra with:
- 100% Cobra API compatibility
- Enhanced terminal UX (colors, spinners, progress bars)
- Built on Charm libraries
- No breaking changes when migrating

## Project Structure

```
bui/
├── commands/          # CLI commands
│   ├── backend/       # Backend generation
│   ├── frontend/      # Frontend generation
│   ├── generate.go    # Generate command router
│   ├── root.go        # Root command
│   ├── start.go       # Start command
│   └── version.go     # Version command
├── utils/             # Utilities and templates
├── version/           # Version management
├── go.mod
├── main.go
└── README.md
```

## Development

```bash
# Install dependencies
go mod tidy

# Build
go build -o bui

# Run
./bui --help
```

## Version

Current version: **0.1.0**

## License

MIT
