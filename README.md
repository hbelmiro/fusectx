# `fusectx`

A CLI tool that recursively resolves a dependency chain of text files and concatenates them into a single output. Perfect for creating context files for AI assistants from hierarchical configuration files.

## Features

- **Hybrid Model**: Supports both inheritance (`extends`) and composition (`includes`)
- **YAML Frontmatter**: Configuration directives defined in YAML frontmatter blocks
- **Circular Dependency Detection**: Prevents infinite loops in dependency chains
- **Multiple Commands**: Build, validate, initialize, and batch processing capabilities

## Installation

```bash
go install github.com/hbelmiro/fusectx/cmd/fusectx@latest
```

Or build from source:

```bash
git clone https://github.com/hbelmiro/fusectx.git
cd fusectx
go build -o fusectx cmd/fusectx/main.go
```

## Quick Start

1. **Initialize a project**:
   ```bash
   fusectx init
   ```

2. **Create a base configuration**:
   ```markdown
   ---
   # base.md
   ---
   # Base Configuration
   
   This is the base content.
   ```

3. **Create a child configuration**:
   ```markdown
   ---
   extends: base.md
   includes:
     - additional.md
   ---
   # Child Configuration
   
   This extends the base and includes additional content.
   ```

4. **Build the final context**:
   ```bash
   fusectx build child.md
   ```

## Resolution Order

When resolving a file, `fusectx` processes content in this order:

1. Content from the recursively resolved `extends` parent chain
2. Content from each file in the `includes` list, in order
3. The content of the current file itself

## Commands

### `fusectx build`

Resolves the full dependency chain and generates the final context.

```bash
fusectx build <source_file> [flags]
```

**Flags:**
- `-o, --output <path>`: Write output to file instead of stdout
- `-s, --silent`: Suppress output messages

**Examples:**
```bash
# Output to stdout
fusectx build config.md

# Output to file
fusectx build config.md -o context.txt

# Silent mode
fusectx build config.md -o context.txt -s
```

### `fusectx init`

Creates a boilerplate `fusectx.md` file to initialize a project.

```bash
fusectx init [directory] [flags]
```

**Flags:**
- `-e, --extends <path>`: Set extends path
- `-i, --includes <path>`: Set includes paths (can be used multiple times)
- `-f, --force`: Overwrite existing file

**Examples:**
```bash
# Initialize in current directory
fusectx init

# Initialize in specific directory
fusectx init ./project

# Initialize with dependencies
fusectx init -e base.md -i common.md -i utils.md
```

### `fusectx validate`

Checks the entire dependency chain for errors without generating output.

```bash
fusectx validate <source_file> [flags]
```

**Flags:**
- `--show-chain`: Display the dependency chain
- `-q, --quiet`: Suppress output messages

**Examples:**
```bash
# Basic validation
fusectx validate config.md

# Show dependency chain
fusectx validate config.md --show-chain

# Quiet validation
fusectx validate config.md -q
```

### `fusectx build-all`

Scans a directory to find and build all leaf project configurations.

```bash
fusectx build-all [directory] [flags]
```

**Flags:**
- `-s, --silent`: Suppress output messages

**Examples:**
```bash
# Build all fusectx.md files in current directory
fusectx build-all

# Build all fusectx.md files in specific directory
fusectx build-all ./projects

# Silent batch build
fusectx build-all ./projects -s
```

## File Format

Files use YAML frontmatter for configuration:

```markdown
---
extends: path/to/parent.md
includes:
  - path/to/include1.md
  - path/to/include2.md
---

# Your Content

Regular markdown content goes here.
```

### Frontmatter Fields

- **`extends`** (string): Path to parent file to inherit from
- **`includes`** (array): List of file paths to include in order

## Examples

### Basic Inheritance

**base.md:**
```markdown
---
---
# Base Configuration

Common settings and documentation.
```

**project.md:**
```markdown
---
extends: base.md
---
# Project Specific

Additional project-specific content.
```

### Complex Dependency Chain

**root.md:**
```markdown
---
---
# Root Documentation
```

**common.md:**
```markdown
---
extends: root.md
---
# Common Utilities
```

**project.md:**
```markdown
---
extends: common.md
includes:
  - apis.md
  - database.md
---
# Project Implementation
```

## Error Handling

- **Circular Dependencies**: Automatically detected and reported
- **Missing Files**: Clear error messages for non-existent dependencies
- **Invalid YAML**: Detailed parsing error information
- **File Permissions**: Appropriate error handling for access issues

## Testing

Run the test suite:

```bash
# Unit tests
go test ./internal/resolver

# Integration tests
go test ./cmd/fusectx

# All tests
go test ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request
