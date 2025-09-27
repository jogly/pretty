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
	"io"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ColorMode controls when colors are used
type ColorMode int

const (
	// ColorAuto automatically detects if colors should be used based on terminal
	ColorAuto ColorMode = iota
	// ColorAlways always uses colors
	ColorAlways
	// ColorNever never uses colors
	ColorNever
)

// Semantic styles using lipgloss
var (
	styleDefault     = lipgloss.NewStyle()
	styleError       = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // red - for errors/invalid
	styleString      = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green - for strings
	styleBoolean     = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // yellow - for booleans
	styleNumber      = lipgloss.NewStyle().Foreground(lipgloss.Color("4")) // blue - for integers
	styleSpecialType = lipgloss.NewStyle().Foreground(lipgloss.Color("5")) // magenta - for special types
	styleFloat       = lipgloss.NewStyle().Foreground(lipgloss.Color("6")) // cyan - for floats
	styleNull        = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // gray - for nil/null
	styleComment     = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // gray - for comments/metadata
	styleField       = lipgloss.NewStyle()                                 // no styling - for field names
)

// Printer configures and performs pretty printing
type Printer struct {
	// MaxWidth is the maximum line width before breaking to multiple lines
	MaxWidth int
	// ColorMode controls when colors are used in output
	ColorMode ColorMode
	// MaxSliceLength is the maximum number of elements to show in slices/arrays
	// If 0, shows all elements (default behavior)
	MaxSliceLength int
	// MaxStringLength is the maximum length for individual strings before truncation
	// If 0, no truncation is applied (default behavior)
	MaxStringLength int
	// styles holds the lipgloss styles for different semantic purposes
	styles struct {
		error       lipgloss.Style // for errors and invalid values
		string      lipgloss.Style // for string values
		boolean     lipgloss.Style // for boolean values
		number      lipgloss.Style // for integer numbers
		float       lipgloss.Style // for floating-point numbers
		specialType lipgloss.Style // for special types like io.ReadCloser
		null        lipgloss.Style // for nil/null values
		comment     lipgloss.Style // for comments and metadata
		field       lipgloss.Style // for field names (struct fields and string map keys)
	}
}

// New creates a new Printer with default options
func New() *Printer {
	p := &Printer{
		MaxWidth:        30,
		ColorMode:       ColorAuto,
		MaxSliceLength:  0, // Show all elements by default
		MaxStringLength: 0, // No string truncation by default
	}

	// Initialize semantic lipgloss styles
	p.styles.error = styleError
	p.styles.string = styleString
	p.styles.boolean = styleBoolean
	p.styles.number = styleNumber
	p.styles.float = styleFloat
	p.styles.specialType = styleSpecialType
	p.styles.null = styleNull
	p.styles.comment = styleComment
	p.styles.field = styleField

	return p
}

// Print formats any input value into a pretty-printed string representation
func (p *Printer) Print(v interface{}) string {
	if v == nil {
		return p.colorize("nil", p.styles.null)
	}

	val := reflect.ValueOf(v)
	return p.formatValue(val, 0)
}

// WithMaxWidth creates a new Printer with the specified maximum width
func (p *Printer) WithMaxWidth(width int) *Printer {
	newP := &Printer{
		MaxWidth:        width,
		ColorMode:       p.ColorMode,
		MaxSliceLength:  p.MaxSliceLength,
		MaxStringLength: p.MaxStringLength,
	}
	newP.styles = p.styles
	return newP
}

// WithColorMode creates a new Printer with the specified color mode
func (p *Printer) WithColorMode(mode ColorMode) *Printer {
	newP := &Printer{
		MaxWidth:        p.MaxWidth,
		ColorMode:       mode,
		MaxSliceLength:  p.MaxSliceLength,
		MaxStringLength: p.MaxStringLength,
	}
	newP.styles = p.styles
	return newP
}

// WithMaxSliceLength creates a new Printer with the specified maximum slice length
func (p *Printer) WithMaxSliceLength(maxLen int) *Printer {
	newP := &Printer{
		MaxWidth:        p.MaxWidth,
		ColorMode:       p.ColorMode,
		MaxSliceLength:  maxLen,
		MaxStringLength: p.MaxStringLength,
	}
	newP.styles = p.styles
	return newP
}

// WithMaxStringLength creates a new Printer with the specified maximum string length
func (p *Printer) WithMaxStringLength(maxLen int) *Printer {
	newP := &Printer{
		MaxWidth:        p.MaxWidth,
		ColorMode:       p.ColorMode,
		MaxSliceLength:  p.MaxSliceLength,
		MaxStringLength: maxLen,
	}
	newP.styles = p.styles
	return newP
}

// shouldUseColors determines if colors should be used based on the color mode
func (p *Printer) shouldUseColors() bool {
	switch p.ColorMode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	case ColorAuto:
		return isTerminal(os.Stdout)
	default:
		return false
	}
}

// isTerminal checks if the given file is a terminal
func isTerminal(f *os.File) bool {
	fileInfo, err := f.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// colorize applies lipgloss styling to text if colors are enabled
func (p *Printer) colorize(text string, style lipgloss.Style) string {
	if !p.shouldUseColors() {
		return text
	}
	return style.Render(text)
}

// Print formats any input value into a pretty-printed string representation using default options
func Print(v interface{}) string {
	return New().Print(v)
}

// formatValue recursively formats a reflect.Value with proper indentation
func (p *Printer) formatValue(val reflect.Value, indent int) string {
	if !val.IsValid() {
		return p.colorize("invalid", p.styles.error)
	}

	// Check if the value implements io.ReadCloser
	if val.IsValid() && val.CanInterface() {
		if _, ok := val.Interface().(io.ReadCloser); ok {
			return p.colorize("<io.ReadCloser>", p.styles.specialType)
		}
	}

	switch val.Kind() {
	case reflect.String:
		str := val.String()
		// Check if string is valid JSON and pretty-print it
		if p.isJSON(str) {
			if prettyJSON := p.formatJSON(str, indent); prettyJSON != "" {
				return prettyJSON
			}
		}
		// Apply string truncation if needed
		truncatedStr := p.truncateString(str)
		return p.colorize(fmt.Sprintf(`"%s"`, truncatedStr), p.styles.string)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return p.colorize(fmt.Sprintf("%d", val.Int()), p.styles.number)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return p.colorize(fmt.Sprintf("%d", val.Uint()), p.styles.number)

	case reflect.Float32, reflect.Float64:
		return p.colorize(fmt.Sprintf("%g", val.Float()), p.styles.float)

	case reflect.Bool:
		return p.colorize(fmt.Sprintf("%t", val.Bool()), p.styles.boolean)

	case reflect.Ptr:
		if val.IsNil() {
			return p.colorize("nil", p.styles.null)
		}
		return p.formatValue(val.Elem(), indent)

	case reflect.Interface:
		if val.IsNil() {
			return p.colorize("nil", p.styles.null)
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

	// Check if slice is too long and should be truncated
	length := val.Len()
	shouldTruncate := p.MaxSliceLength > 0 && length > p.MaxSliceLength

	if shouldTruncate {
		return p.formatTruncatedSlice(val, indent, length)
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

// formatTruncatedSlice formats a long slice by showing first few, last few, and a summary
func (p *Printer) formatTruncatedSlice(val reflect.Value, indent int, totalLength int) string {
	maxLen := p.MaxSliceLength
	showCount := maxLen / 2 // Show half at beginning, half at end
	if showCount < 1 {
		showCount = 1
	}

	var parts []string
	nextIndent := indent + 1
	indentStr := strings.Repeat("  ", nextIndent)

	// Show first elements
	for i := 0; i < showCount && i < totalLength; i++ {
		elem := p.formatValue(val.Index(i), nextIndent)
		parts = append(parts, indentStr+elem)
	}

	// Add truncation indicator
	omittedCount := totalLength - (2 * showCount)
	if omittedCount > 0 {
		truncMsg := fmt.Sprintf("... %d more elements ...", omittedCount)
		parts = append(parts, indentStr+p.colorize(truncMsg, p.styles.comment))
	}

	// Show last elements
	startIdx := totalLength - showCount
	if startIdx < showCount {
		startIdx = showCount // Avoid overlap
	}
	for i := startIdx; i < totalLength; i++ {
		elem := p.formatValue(val.Index(i), nextIndent)
		parts = append(parts, indentStr+elem)
	}

	// Add summary comment
	summary := fmt.Sprintf("// Total length: %d", totalLength)
	parts = append(parts, indentStr+p.colorize(summary, p.styles.comment))

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
		keyStr := p.formatMapKey(key)
		mapValue := val.MapIndex(key)

		// Check if we should omit struct name when key matches struct type
		var valueStr string
		if key.Kind() == reflect.String {
			// Handle both direct structs and interface-wrapped structs
			actualValue := mapValue
			if mapValue.Kind() == reflect.Interface && !mapValue.IsNil() {
				actualValue = mapValue.Elem()
			}

			if actualValue.Kind() == reflect.Struct {
				structTypeName := actualValue.Type().Name()
				if key.String() == structTypeName {
					// Key matches struct name, format struct without type name
					valueStr = p.formatStructWithName(actualValue, 0, false)
				} else {
					valueStr = p.formatValue(mapValue, 0)
				}
			} else {
				valueStr = p.formatValue(mapValue, 0) // Use 0 indent for single line
			}
		} else {
			valueStr = p.formatValue(mapValue, 0) // Use 0 indent for single line
		}

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
		keyStr := p.formatMapKey(key)
		mapValue := val.MapIndex(key)

		// Check if we should omit struct name when key matches struct type
		var valueStr string
		if key.Kind() == reflect.String {
			// Handle both direct structs and interface-wrapped structs
			actualValue := mapValue
			if mapValue.Kind() == reflect.Interface && !mapValue.IsNil() {
				actualValue = mapValue.Elem()
			}

			if actualValue.Kind() == reflect.Struct {
				structTypeName := actualValue.Type().Name()
				if key.String() == structTypeName {
					// Key matches struct name, format struct without type name
					valueStr = p.formatStructWithName(actualValue, nextIndent, false)
				} else {
					valueStr = p.formatValue(mapValue, nextIndent)
				}
			} else {
				valueStr = p.formatValue(mapValue, nextIndent)
			}
		} else {
			valueStr = p.formatValue(mapValue, nextIndent)
		}

		parts = append(parts, fmt.Sprintf("%s%s: %s", indentStr, keyStr, valueStr))
	}

	return fmt.Sprintf("{\n%s\n%s}", strings.Join(parts, ",\n"), strings.Repeat("  ", indent))
}

// formatMapKey formats a map key, treating string keys like struct field names
func (p *Printer) formatMapKey(key reflect.Value) string {
	// If the key is a string, format it like a struct field (no quotes, no coloring)
	if key.Kind() == reflect.String {
		str := key.String()
		// Apply string truncation if needed, but no quotes or styling
		truncatedStr := p.truncateString(str)
		return p.colorize(truncatedStr, p.styles.field)
	} else if key.Kind() == reflect.Struct {
		return p.formatStructWithName(key, 0, false)
	}

	// For non-string keys, use the regular formatting
	return p.formatValue(key, 0)
}

// formatStruct formats structs
func (p *Printer) formatStruct(val reflect.Value, indent int) string {
	return p.formatStructWithName(val, indent, true)
}

// formatStructWithName formats structs with optional struct name
func (p *Printer) formatStructWithName(val reflect.Value, indent int, includeTypeName bool) string {
	typ := val.Type()
	if val.NumField() == 0 {
		if includeTypeName {
			return fmt.Sprintf("%s{}", typ.Name())
		}
		return "{}"
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

		// Check if field name matches struct type name and omit struct name if so
		var fieldStr string
		if fieldVal.Kind() == reflect.Struct && field.Name == fieldVal.Type().Name() {
			fieldStr = p.formatStructWithName(fieldVal, 0, false) // Omit struct name
		} else {
			fieldStr = p.formatValue(fieldVal, 0) // Use 0 indent for single line
		}

		fieldPart := fmt.Sprintf("%s: %s", field.Name, fieldStr)
		exportedFields = append(exportedFields, fieldPart)
		singleLineParts = append(singleLineParts, fieldPart)
	}

	if len(exportedFields) == 0 {
		if includeTypeName {
			return fmt.Sprintf("%s{}", typ.Name())
		}
		return "{}"
	}

	// Try single line format first
	var singleLine string
	if includeTypeName {
		singleLine = fmt.Sprintf("%s{%s}", typ.Name(), strings.Join(singleLineParts, ", "))
	} else {
		singleLine = fmt.Sprintf("{%s}", strings.Join(singleLineParts, ", "))
	}

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

		// Check if field name matches struct type name and omit struct name if so
		var fieldStr string
		if fieldVal.Kind() == reflect.Struct && field.Name == fieldVal.Type().Name() {
			fieldStr = p.formatStructWithName(fieldVal, nextIndent, false) // Omit struct name
		} else {
			fieldStr = p.formatValue(fieldVal, nextIndent)
		}

		parts = append(parts, fmt.Sprintf("%s%s: %s", indentStr, field.Name, fieldStr))
	}

	if includeTypeName {
		return fmt.Sprintf("%s{\n%s\n%s}", typ.Name(), strings.Join(parts, ",\n"), strings.Repeat("  ", indent))
	} else {
		return fmt.Sprintf("{\n%s\n%s}", strings.Join(parts, ",\n"), strings.Repeat("  ", indent))
	}
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

// isJSON checks if a string is valid JSON
func (p *Printer) isJSON(str string) bool {
	if len(str) < 2 {
		return false
	}

	// Quick check for JSON-like structure
	trimmed := strings.TrimSpace(str)
	if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {

		var js json.RawMessage
		return json.Unmarshal([]byte(str), &js) == nil
	}
	return false
}

// formatJSON formats a JSON string with proper indentation and colors
func (p *Printer) formatJSON(jsonStr string, indent int) string {
	var parsed interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return ""
	}

	// Use our own formatter to format the parsed JSON with colors
	return p.formatValue(reflect.ValueOf(parsed), indent)
}

// truncateString truncates a string with center ellipses if it exceeds MaxStringLength
func (p *Printer) truncateString(str string) string {
	if p.MaxStringLength <= 0 || len(str) <= p.MaxStringLength {
		return str
	}

	maxLen := p.MaxStringLength
	ellipses := "..."

	// Need at least 4 characters to show something meaningful (at least 1 char + ... + 1 char)
	if maxLen < 4 {
		return str[:maxLen]
	}

	// Calculate how much space we have for actual content
	contentLen := maxLen - len(ellipses)
	leftLen := contentLen / 2
	rightLen := contentLen - leftLen

	// Handle edge case where string is shorter than expected after calculation
	if leftLen+rightLen >= len(str) {
		return str
	}

	return str[:leftLen] + ellipses + str[len(str)-rightLen:]
}
