// Package pretty provides utilities for formatting any input into pretty-printed strings.
//
// Basic usage:
//
//	pretty.Print(data)
//
// With custom width:
//
//	pp := &pretty.Printer{MaxWidth: 50}
//	pp.Print(data)
//
// Or using the fluent API:
//
//	pretty.New().WithMaxWidth(50).Print(data)
package pretty

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Printer configures and performs pretty printing
type Printer struct {
	// MaxWidth is the maximum line width before breaking to multiple lines
	MaxWidth int
}

// New creates a new Printer with default options
func New() *Printer {
	return &Printer{
		MaxWidth: 30,
	}
}

// Print formats any input value into a pretty-printed string representation
func (p *Printer) Print(v interface{}) string {
	if v == nil {
		return "nil"
	}

	val := reflect.ValueOf(v)
	return p.formatValue(val, 0)
}

// WithMaxWidth creates a new Printer with the specified maximum width
func (p *Printer) WithMaxWidth(width int) *Printer {
	return &Printer{
		MaxWidth: width,
	}
}

// Print formats any input value into a pretty-printed string representation using default options
func Print(v interface{}) string {
	return New().Print(v)
}

// formatValue recursively formats a reflect.Value with proper indentation
func (p *Printer) formatValue(val reflect.Value, indent int) string {
	if !val.IsValid() {
		return "invalid"
	}

	switch val.Kind() {
	case reflect.String:
		return fmt.Sprintf(`"%s"`, val.String())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", val.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", val.Uint())

	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", val.Float())

	case reflect.Bool:
		return fmt.Sprintf("%t", val.Bool())

	case reflect.Ptr:
		if val.IsNil() {
			return "nil"
		}
		return p.formatValue(val.Elem(), indent)

	case reflect.Interface:
		if val.IsNil() {
			return "nil"
		}
		return p.formatValue(val.Elem(), indent)

	case reflect.Slice, reflect.Array:
		return p.formatSlice(val, indent)

	case reflect.Map:
		return p.formatMap(val, indent)

	case reflect.Struct:
		return p.formatStruct(val, indent)
	case reflect.Chan:
		return p.formatChan(val)
	default:
		// Fallback to JSON for complex types
		if data, err := json.MarshalIndent(val.Interface(), strings.Repeat("  ", indent), "  "); err == nil {
			return string(data)
		}
		return fmt.Sprintf("%+v", val.Interface())
	}
}

// formatSlice formats slices and arrays
func (p *Printer) formatSlice(val reflect.Value, indent int) string {
	if val.Len() == 0 {
		return "[]"
	}

	// Try single line format first
	var singleLineParts []string
	for i := 0; i < val.Len(); i++ {
		elem := p.formatValue(val.Index(i), 0) // Use 0 indent for single line
		singleLineParts = append(singleLineParts, elem)
	}

	singleLine := fmt.Sprintf("[%s]", strings.Join(singleLineParts, ", "))

	// Use single line if it's within the max width
	if len(singleLine) <= p.MaxWidth {
		return singleLine
	}

	// Fall back to multi-line format
	var parts []string
	nextIndent := indent + 1
	indentStr := strings.Repeat("  ", nextIndent)

	for i := 0; i < val.Len(); i++ {
		elem := p.formatValue(val.Index(i), nextIndent)
		parts = append(parts, indentStr+elem)
	}

	return fmt.Sprintf("[\n%s\n%s]", strings.Join(parts, ",\n"), strings.Repeat("  ", indent))
}

// formatMap formats maps
func (p *Printer) formatMap(val reflect.Value, indent int) string {
	if val.Len() == 0 {
		return "{}"
	}

	// Sort keys for consistent output
	keys := val.MapKeys()
	p.sortMapKeys(keys)

	// Try single line format first
	var singleLineParts []string
	for _, key := range keys {
		keyStr := p.formatValue(key, 0)
		valueStr := p.formatValue(val.MapIndex(key), 0) // Use 0 indent for single line
		singleLineParts = append(singleLineParts, fmt.Sprintf("%s: %s", keyStr, valueStr))
	}

	singleLine := fmt.Sprintf("{%s}", strings.Join(singleLineParts, ", "))

	// Use single line if it's within the max width
	if len(singleLine) <= p.MaxWidth {
		return singleLine
	}

	// Fall back to multi-line format
	var parts []string
	nextIndent := indent + 1
	indentStr := strings.Repeat("  ", nextIndent)

	for _, key := range keys {
		keyStr := p.formatValue(key, 0)
		valueStr := p.formatValue(val.MapIndex(key), nextIndent)
		parts = append(parts, fmt.Sprintf("%s%s: %s", indentStr, keyStr, valueStr))
	}

	return fmt.Sprintf("{\n%s\n%s}", strings.Join(parts, ",\n"), strings.Repeat("  ", indent))
}

// formatStruct formats structs
func (p *Printer) formatStruct(val reflect.Value, indent int) string {
	typ := val.Type()
	if val.NumField() == 0 {
		return fmt.Sprintf("%s{}", typ.Name())
	}

	// Collect exported fields first
	var exportedFields []string
	var singleLineParts []string

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldVal := val.Field(i)
		fieldStr := p.formatValue(fieldVal, 0) // Use 0 indent for single line
		fieldPart := fmt.Sprintf("%s: %s", field.Name, fieldStr)
		exportedFields = append(exportedFields, fieldPart)
		singleLineParts = append(singleLineParts, fieldPart)
	}

	if len(exportedFields) == 0 {
		return fmt.Sprintf("%s{}", typ.Name())
	}

	// Try single line format first
	singleLine := fmt.Sprintf("%s{%s}", typ.Name(), strings.Join(singleLineParts, ", "))

	// Use single line if it's within the max width
	if len(singleLine) <= p.MaxWidth {
		return singleLine
	}

	// Fall back to multi-line format
	var parts []string
	nextIndent := indent + 1
	indentStr := strings.Repeat("  ", nextIndent)

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldVal := val.Field(i)
		fieldStr := p.formatValue(fieldVal, nextIndent)
		parts = append(parts, fmt.Sprintf("%s%s: %s", indentStr, field.Name, fieldStr))
	}

	return fmt.Sprintf("%s{\n%s\n%s}", typ.Name(), strings.Join(parts, ",\n"), strings.Repeat("  ", indent))
}

func (p *Printer) formatChan(val reflect.Value) string {
	dir := val.Type().ChanDir()
	elemType := val.Type().Elem().String()

	switch dir {
	case reflect.RecvDir:
		return fmt.Sprintf("<-chan %s", elemType)
	case reflect.SendDir:
		return fmt.Sprintf("chan<- %s", elemType)
	case reflect.BothDir:
		return fmt.Sprintf("chan %s", elemType)
	default:
		panic(fmt.Sprintf("invalid channel direction: %s", dir))
	}
}

// sortMapKeys sorts map keys for consistent output
func (p *Printer) sortMapKeys(keys []reflect.Value) {
	sort.Slice(keys, func(i, j int) bool {
		// Convert keys to strings for comparison
		iStr := p.keyToString(keys[i])
		jStr := p.keyToString(keys[j])
		return iStr < jStr
	})
}

// keyToString converts a map key to a string for sorting
func (p *Printer) keyToString(key reflect.Value) string {
	switch key.Kind() {
	case reflect.String:
		return key.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", key.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", key.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", key.Float())
	case reflect.Bool:
		return fmt.Sprintf("%t", key.Bool())
	default:
		// Fallback to formatted value for other types
		return p.formatValue(key, 0)
	}
}
