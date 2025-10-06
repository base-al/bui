# Mamba Features Integration

Bui CLI uses [Mamba](https://github.com/base-go/mamba) - a modern CLI framework with enhanced UX features.

## Features Implemented

### 1. Colored Output

All output uses Mamba's colored print methods for better readability:

- `cmd.PrintSuccess()` - ✅ Green success messages
- `cmd.PrintError()` - ❌ Red error messages
- `cmd.PrintInfo()` - 💡 Blue informational messages
- `cmd.PrintWarning()` - ⚠️  Yellow warnings

### 2. Loading Spinners

Long-running operations show animated spinners (hidden in verbose mode):

```bash
# Shows spinner
bui g product name:string price:float
# Output: ⠋ Generating backend module...

# Verbose shows detailed output
bui g product name:string price:float -v
# Output:
# Created directory: app/models
# Created directory: app/products
# Formatting generated files...
# Running go mod tidy...
```

### 3. Verbose Flag

Global `-v` or `--verbose` flag available on all commands:

```bash
bui new my-project              # Clean output with spinners
bui new my-project -v           # Detailed progress output

bui g product name:string       # Shows spinner
bui g product name:string -v    # Shows all steps

bui d product                   # Minimal feedback
bui d product -v                # Detailed deletion info
```

## Commands Using Mamba Features

### ✅ Fully Integrated
- `bui generate` - Spinners + colored output + verbose mode
- `bui generate backend` - Colored output + verbose mode
- `bui new` - Spinners for cloning + colored output + verbose mode
- `bui destroy` - Colored output
- `bui build` - Spinners for building
- `bui update` - Spinners for downloading
- `bui dev` - Colored output

### 🚧 Partially Integrated
- `bui start` - Uses plain fmt (can be enhanced)
- `bui version` - Uses plain fmt (can be enhanced)

## UX Improvements

### Before Mamba
```
Generating backend module...
Error creating directory app/models: permission denied
Successfully generated product module
```

### After Mamba
```
⠋ Generating backend module...
✅ Successfully generated product module (backend + frontend)

# Or with -v flag:
Created directory: app/models
Created directory: app/products
Formatting generated files...
✅ Added module to app/init.go
Running go mod tidy...
✅ Generated backend module: product
```

## Developer Notes

### Adding Spinners to New Commands

```go
import "github.com/base-go/mamba/pkg/spinner"

if !Verbose {
    err := spinner.WithSpinner("Processing...", func() error {
        // Long running operation here
        return doWork()
    })
    if err != nil {
        cmd.PrintError("Operation failed")
        return
    }
    cmd.PrintSuccess("Operation completed")
} else {
    cmd.PrintInfo("Processing...")
    if err := doWork(); err != nil {
        cmd.PrintError("Operation failed")
        return
    }
    cmd.PrintSuccess("Operation completed")
}
```

### Using Colored Output

```go
// Success (green)
cmd.PrintSuccess("Module generated successfully")

// Error (red)
cmd.PrintError("Failed to create directory")

// Info (blue)
cmd.PrintInfo("Starting backend server...")

// Warning (yellow)
cmd.PrintWarning("goimports not found, installing...")
```

### Accessing Verbose Flag

In command files:
```go
import "github.com/base-al/bui/commands"

if commands.Verbose {
    // Show detailed output
    cmd.PrintInfo("Detailed step information...")
}
```

In subcommand packages (backend/frontend):
```go
// In globals.go
var Verbose *bool

// In parent command
backend.Verbose = &commands.Verbose

// In subcommand
if Verbose != nil && *Verbose {
    cmd.PrintInfo("Detailed output...")
}
```

## Benefits

1. **Professional Appearance** - Modern CLI UX with spinners and colors
2. **Better Feedback** - Clear visual distinction between success/error/info
3. **Flexibility** - Users can choose minimal or detailed output
4. **Consistent Experience** - All commands use the same patterns
5. **Reduced Clutter** - Non-verbose mode hides technical details

## Future Enhancements

- [ ] Add progress bars for file operations
- [ ] Add interactive prompts with Mamba's bubble tea integration
- [ ] Add table formatting for listing resources
- [ ] Add command completion suggestions
- [ ] Add emoji support toggle
