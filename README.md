# Pretty

A Go library for pretty-printing any data structure with optional ANSI color support.

## Features

- Pretty-print any Go data type (structs, slices, maps, primitives, etc.)
- Configurable line width for automatic multi-line formatting
- ANSI color support with automatic terminal detection
- Clean, readable output with proper indentation
- Fluent API for easy configuration

## Installation

```bash
go get github.com/jogly/pretty
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/jogly/pretty"
)

func main() {
    data := map[string]interface{}{
        "name": "Alice",
        "age": 30,
        "active": true,
        "balance": 123.45,
        "tags": []string{"user", "premium"},
    }

    fmt.Println(pretty.Print(data))
}
```

### Custom Configuration

```go
// Custom width
pp := &pretty.Printer{MaxWidth: 50}
fmt.Println(pp.Print(data))

// Fluent API
output := pretty.New().
    WithMaxWidth(80).
    WithColorMode(pretty.ColorAlways).
    Print(data)
```

### Color Options

```go
// Auto-detect terminal support (default)
printer := pretty.New().WithColorMode(pretty.ColorAuto)

// Always use colors
printer := pretty.New().WithColorMode(pretty.ColorAlways)

// Never use colors
printer := pretty.New().WithColorMode(pretty.ColorNever)
```

## Color Scheme

When colors are enabled:
- **Strings**: Green
- **Numbers**: Blue
- **Booleans**: Yellow
- **Floats**: Cyan
- **nil/invalid**: Gray
- **Errors**: Red

## Examples

### Simple Data Types

```go
pretty.Print("hello")     // "hello" (green if colors enabled)
pretty.Print(42)          // 42 (blue if colors enabled)
pretty.Print(true)        // true (yellow if colors enabled)
pretty.Print(nil)         // nil (gray if colors enabled)
```

### Complex Structures

```go
type Person struct {
    Name    string
    Age     int
    Hobbies []string
}

person := Person{
    Name:    "John",
    Age:     30,
    Hobbies: []string{"reading", "coding"},
}

fmt.Println(pretty.Print(person))
```

Output:
```
Person{
  Name: "John",
  Age: 30,
  Hobbies: [
    "reading",
    "coding"
  ]
}
```

## API Reference

### Types

```go
type ColorMode int

const (
    ColorAuto   ColorMode = iota  // Auto-detect terminal
    ColorAlways                   // Always use colors
    ColorNever                    // Never use colors
)

type Printer struct {
    MaxWidth  int       // Maximum line width
    ColorMode ColorMode // Color configuration
}
```

### Functions

```go
// Create new printer with defaults
func New() *Printer

// Print with default settings
func Print(v interface{}) string

// Configure printer
func (p *Printer) WithMaxWidth(width int) *Printer
func (p *Printer) WithColorMode(mode ColorMode) *Printer
func (p *Printer) Print(v interface{}) string
```

## License

MIT