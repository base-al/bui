# Mamba Features Integration

Bui CLI uses [Mamba](https://github.com/base-go/mamba) - a modern, drop-in replacement for Cobra with enhanced UX features.

## Available Mamba Features

### 1. Printing Methods

Mamba provides multiple styled print methods:

- `cmd.PrintSuccess()` - ✅ Green success messages
- `cmd.PrintError()` - ❌ Red error messages
- `cmd.PrintInfo()` - 💡 Blue informational messages
- `cmd.PrintWarning()` - ⚠️  Yellow warnings
- `cmd.PrintHeader()` - Section headers
- `cmd.PrintBullet()` - Bullet point items
- `cmd.PrintCode()` - Code snippet display
- `cmd.PrintBox()` - Message in a box

### 2. Loading Spinners

Long-running operations show animated spinners:

```go
spinner.WithSpinner("Processing...", func() error {
    // Long running operation
    return doWork()
})
```

Example output:
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
```

### 3. Progress Bars

Track progress for batch operations:

```go
progress.WithProgress("Processing files...", totalCount, func(update func(int)) error {
    for i, file := range files {
        processFile(file)
        update(i + 1)
    }
    return nil
})
```

### 4. Interactive Prompts

Collect user input interactively:

```go
// Text input
name := cmd.AskString("Project name:")

// Yes/No confirmation
confirmed := cmd.AskConfirm("Continue?")

// Single selection
choice := cmd.AskSelect("Choose option:", []string{"A", "B", "C"})

// Multiple selection
selected := cmd.AskMultiSelect("Select features:", []string{"Auth", "API", "DB"})
```

### 5. Styling Methods

Custom styling for terminal output:

```go
mamba.Success("Done!")    // Green styled text
mamba.Error("Failed!")    // Red styled text
mamba.Bold("Important")   // Bold text
mamba.Code("go build")    // Code styled text
mamba.Header("Section")   // Header styled text
mamba.Box("Message")      // Boxed message
```

### 6. Verbose Mode

Global `-v` or `--verbose` flag on all commands:

```bash
bui new my-project              # Clean output with spinners
bui new my-project -v           # Detailed progress output

bui g product name:string       # Shows spinner
bui g product name:string -v    # Shows all steps
```

## Current Implementation Status

### ✅ Fully Integrated Commands

These commands use Mamba's colored output, spinners, and interactive features:

- `bui generate` - Colored output + verbose mode
- `bui generate backend` - Colored output + verbose mode
- `bui generate frontend` - Colored output + verbose mode
- `bui new` - Spinners + colored output + headers + bullets + verbose mode
- `bui destroy` - Interactive confirmation + colored output + headers + bullets + verbose mode
- `bui destroy backend` - Interactive confirmation + colored output
- `bui destroy frontend` - Interactive confirmation + colored output
- `bui build` - Spinners + colored output + headers + bullets
- `bui build backend` - Spinners + colored output
- `bui build frontend` - Spinners + colored output
- `bui upgrade` - Spinners + colored output + headers + bullets
- `bui dev` - Colored output
- `bui start` - Spinners + colored output + headers + verbose mode
- `bui version` - Colored output + update warnings

## Why Mamba vs Cobra?

Bui migrated from Cobra to Mamba to get modern CLI UX features that Cobra doesn't provide:

| Feature | Cobra | Mamba |
|---------|-------|-------|
| Command structure | ✅ Yes | ✅ Yes |
| Subcommands & nesting | ✅ Yes | ✅ Yes |
| POSIX flags (pflag) | ✅ Yes | ✅ Yes |
| Lifecycle hooks | ✅ Yes | ✅ Yes |
| Argument validation | ✅ Yes | ✅ Yes |
| Command aliases | ✅ Yes | ✅ Yes |
| Help generation | ✅ Yes | ✅ Yes (Enhanced & Styled) |
| Shell autocomplete | ✅ Yes | ⏳ No (planned) |
| Intelligent suggestions | ✅ Yes | ⏳ No (planned) |
| Man page generation | ✅ Yes | ⏳ No (planned) |
| Viper integration | 🔧 Optional | ⏳ No (planned) |
| **Colored output** | ❌ No | ✅ **Yes** |
| **Loading spinners** | ❌ No | ✅ **Yes** |
| **Progress bars** | ❌ No | ✅ **Yes** |
| **Interactive prompts** | ❌ No | ✅ **Yes** |
| **Modern terminal UX** | ❌ No | ✅ **Yes** |

**Key Benefits**:
- ✅ 100% Cobra API compatibility - no breaking changes when migrating
- ✅ Enhanced UX with spinners, progress bars, and colored output
- ✅ Built on Charm libraries (Bubble Tea, Lipgloss) for modern terminal styling
- ✅ Professional CLI experience for end users
- ✅ Reduces terminal clutter with verbose mode

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

## Future Enhancement Ideas

These Mamba features are available but not yet implemented in Bui:

### Progress Bars
- [ ] Show progress when generating multiple modules
- [ ] Track file copying progress during `bui new`
- [ ] Display build progress for large projects

### Interactive Prompts
- [ ] `bui new` - Interactive project setup wizard
  - Ask for project name
  - Select features to include (auth, media, etc.)
  - Choose database (PostgreSQL, MySQL, SQLite)
- [ ] `bui generate` - Interactive field builder
  - Add fields one by one with prompts
  - Select field types from list
  - Configure relationships interactively
- [ ] `bui destroy` - Confirmation prompt before deletion

### Enhanced Styling
- [ ] Use `PrintHeader()` for section headers
- [ ] Use `PrintBullet()` for file lists
- [ ] Use `PrintCode()` for showing generated code snippets
- [ ] Use `PrintBox()` for important warnings

### Better Output Formatting
- [ ] Table output for listing modules/dependencies
- [ ] Formatted diffs when updating files
- [ ] Syntax-highlighted code previews

## Examples of Potential Future Features

### Interactive Module Generation
```bash
$ bui generate --interactive
? Module name: product
? Add fields interactively? Yes
? Field name: name
? Field type: (Use arrow keys)
  > string
    text
    int
    float
    bool
? Add another field? Yes
...
✅ Generated product module with 5 fields
```

### Progress Bar for Batch Operations
```bash
$ bui new my-project
⠋ Cloning repositories...
████████████████████ 50% Cloning backend template...
████████████████████████████████████████ 100% Complete!
✅ Project created successfully
```

### Styled Help Output
```bash
$ bui generate --help
Generate Modules

Usage:
  bui generate [module] [field:type...]

Examples:
  bui g product name:string price:float
  bui g backend user email:string
  bui g frontend dashboard stats:int

Available Commands:
  backend   Generate backend module only
  frontend  Generate frontend module only
```
