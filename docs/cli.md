# CLI Reference

## Installation

```bash
go install github.com/RodolfoBonis/deco/cmd/deco@latest
```

## Commands

### init

Initialize a new deco project:

```bash
deco init
```

Creates:
- `.deco.yaml` - Configuration file
- `.gitignore` - Git ignore file (if not exists)

### generate

Generate the decorators initialization code:

```bash
deco generate
```

**Options:**
- `--config <file>` - Use custom configuration file (default: .deco.yaml)
- `--verbose` - Enable verbose output
- `--watch` - Watch for file changes and regenerate

### dev

Start development mode with hot reload:

```bash
deco dev
```

**Features:**
- File watching and auto-regeneration
- Hot reload for development
- Verbose logging

### build

Build for production:

```bash
deco build
```

**Options:**
- `--minify` - Minify generated code
- `--validate` - Validate generated code

### validate

Validate decorators and configuration:

```bash
deco validate
```

**Options:**
- `--strict` - Strict validation mode
- `--verbose` - Verbose output

## Configuration

The CLI uses `.deco.yaml` configuration file:

```yaml
handlers:
  include:
    - "handlers/**/*.go"
    - "api/**/*.go"
  exclude:
    - "**/*_test.go"

generation:
  output: ".deco/init_decorators.go"
  package: "deco"

dev:
  watch: true
  hot_reload: true

prod:
  minify: true
  validate: true
```

## Examples

### Basic Usage

```bash
# Initialize project
deco init

# Generate code
deco generate

# Development mode
deco dev

# Production build
deco build
```

### With Options

```bash
# Use custom configuration
deco generate --config custom.deco.yaml

# Verbose generation
deco generate --verbose

# Watch mode
deco generate --watch
```

### Validation

```bash
# Basic validation
deco validate

# Strict validation
deco validate --strict
```

## Next Steps

- **[Installation Guide](./installation.md)** - Setup instructions
- **[Usage Guide](./usage.md)** - How to use decorators
- **[API Reference](./api.md)** - Complete API documentation
