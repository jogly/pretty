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
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

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

const (
	defaultWidth = 100
)

// Semantic styles using lipgloss
var (
	styleError       = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))  // red - for errors/invalid
	styleString      = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))  // green - for strings
	styleBoolean     = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))  // yellow - for booleans
	styleNumber      = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))  // blue - for integers
	styleSpecialType = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))  // magenta - for special types
	styleFloat       = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))  // cyan - for floats
	styleNull        = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // gray - for nil/null
	styleComment     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // gray - for comments/metadata
	styleTime        = lipgloss.NewStyle().Foreground(lipgloss.Color("13")) // bright magenta - for times
	stylePointer     = lipgloss.NewStyle().Foreground(lipgloss.Color("88"))
	styleField       = lipgloss.NewStyle() // no styling - for field names

	pointerGamut = []lipgloss.Style{
		lipgloss.NewStyle().Foreground(lipgloss.Color("196")), // Red
		lipgloss.NewStyle().Foreground(lipgloss.Color("208")), // Orange
		lipgloss.NewStyle().Foreground(lipgloss.Color("226")), // Yellow
		lipgloss.NewStyle().Foreground(lipgloss.Color("51")),  // Cyan
		lipgloss.NewStyle().Foreground(lipgloss.Color("135")), // Purple
		lipgloss.NewStyle().Foreground(lipgloss.Color("170")), // Pink
		lipgloss.NewStyle().Foreground(lipgloss.Color("129")), // Blue
		lipgloss.NewStyle().Foreground(lipgloss.Color("204")), // Teal
		lipgloss.NewStyle().Foreground(lipgloss.Color("124")), // Green
	}
)

var (
	timeType = reflect.TypeOf(time.Time{})

	Default = New()
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
	// MaxKeysInline is the maximum number of keys to show in maps & structs
	// inline before breaking to multiple lines. This allows maps and structs to
	// be multi-lined before the overall max width is reached.
	MaxKeysInline int
	// Margin adds space around the output
	// If 0, no margin is applied (default behavior)
	Margin [4]int

	// Styles holds the lipgloss Styles for different semantic purposes
	Styles struct {
		Error       lipgloss.Style // for errors and invalid values
		String      lipgloss.Style // for string values
		Boolean     lipgloss.Style // for boolean values
		Number      lipgloss.Style // for integer numbers
		Float       lipgloss.Style // for floating-point numbers
		SpecialType lipgloss.Style // for special types like io.ReadCloser
		Time        lipgloss.Style // for time values
		Null        lipgloss.Style // for nil/null values
		Comment     lipgloss.Style // for comments and metadata
		Field       lipgloss.Style // for field names (struct fields and string map keys)
		Pointer     lipgloss.Style // for pointers
	}

	visited map[uintptr]bool
	cycled  map[uintptr]bool
}

// New creates a new Printer with default options
func New() *Printer {
	p := &Printer{
		MaxWidth:        defaultWidth,
		ColorMode:       ColorAuto,
		MaxSliceLength:  20,
		MaxStringLength: 0, // No string truncation by default
		Margin:          [4]int{0, 0, 0, 0},
	}

	// Initialize semantic lipgloss styles
	p.Styles.Error = styleError
	p.Styles.String = styleString
	p.Styles.Boolean = styleBoolean
	p.Styles.Number = styleNumber
	p.Styles.Float = styleFloat
	p.Styles.SpecialType = styleSpecialType
	p.Styles.Time = styleTime
	p.Styles.Null = styleNull
	p.Styles.Comment = styleComment
	p.Styles.Field = styleField
	p.Styles.Pointer = stylePointer

	return p
}

// Print formats any input value into a pretty-printed string representation
func (p *Printer) Print(v interface{}) string {
	if v == nil {
		return p.colorize("nil", p.Styles.Null)
	}

	val := reflect.ValueOf(v)

	p.visited = make(map[uintptr]bool)
	p.cycled = make(map[uintptr]bool)
	defer clear(p.visited)

	if p.Margin[0] != 0 || p.Margin[1] != 0 || p.Margin[2] != 0 || p.Margin[3] != 0 {
		style := lipgloss.NewStyle().Margin(p.Margin[0], p.Margin[1], p.Margin[2], p.Margin[3])
		return style.Render(p.formatValue(val, 0))
	}

	return p.formatValue(val, 0)
}

// copyPrinter creates a copy of the printer with optional field overrides
func (p *Printer) copyPrinter() *Printer {
	newP := *p // Shallow copy
	return &newP
}

// WithMaxWidth creates a new Printer with the specified maximum width
func (p *Printer) WithMaxWidth(width int) *Printer {
	newP := p.copyPrinter()
	newP.MaxWidth = width
	return newP
}

// WithColorMode creates a new Printer with the specified color mode
func (p *Printer) WithColorMode(mode ColorMode) *Printer {
	newP := p.copyPrinter()
	newP.ColorMode = mode
	return newP
}

// WithMaxSliceLength creates a new Printer with the specified maximum slice length
func (p *Printer) WithMaxSliceLength(maxLen int) *Printer {
	newP := p.copyPrinter()
	newP.MaxSliceLength = maxLen
	return newP
}

// WithMaxStringLength creates a new Printer with the specified maximum string length
func (p *Printer) WithMaxStringLength(maxLen int) *Printer {
	newP := p.copyPrinter()
	newP.MaxStringLength = maxLen
	return newP
}

func (p *Printer) WithMargin(margin ...int) *Printer {
	switch len(margin) {
	case 1:
		p.Margin[0] = margin[0]
		p.Margin[1] = margin[0]
		p.Margin[2] = margin[0]
		p.Margin[3] = margin[0]
	case 2:
		p.Margin[0] = margin[0]
		p.Margin[1] = margin[1]
		p.Margin[2] = margin[0]
		p.Margin[3] = margin[1]
	case 3:
		p.Margin[0] = margin[0]
		p.Margin[1] = margin[1]
		p.Margin[2] = margin[2]
		p.Margin[3] = margin[0]
	case 4:
		p.Margin[0] = margin[0]
		p.Margin[1] = margin[1]
		p.Margin[2] = margin[2]
		p.Margin[3] = margin[3]
	}
	return p
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

// compoundFormatter handles single-line vs multi-line formatting for compound types
type compoundFormatter struct {
	p             *Printer
	openBrace     string
	closeBrace    string
	typeName      string
	singleItems   []string
	multiItems    []string
	indent        int
	currentWidth  int  // Running tally of visible width
	exceedsWidth  bool // Early escape flag when width is exceeded
	padBraces     bool // Whether to pad the braces with spaces in single-line format
	maxKeysInline int
}

// newCompoundFormatter creates a new compound formatter
func (p *Printer) newCompoundFormatter(openBrace, closeBrace, typeName string, indent int, padBraces bool, maxKeysInline int) *compoundFormatter {
	cf := &compoundFormatter{
		p:             p,
		openBrace:     openBrace,
		closeBrace:    closeBrace,
		typeName:      typeName,
		indent:        indent,
		padBraces:     padBraces,
		maxKeysInline: maxKeysInline,
	}

	// Initialize width with opening elements
	padded := 0
	if padBraces {
		padded = 1
	}
	cf.currentWidth = lipgloss.Width(typeName+openBrace) + padded

	return cf
}

// addItem adds an item to both single and multi-line formats
func (cf *compoundFormatter) addItem(singleItem, multiItem string) {
	// Early escape optimization: if we already exceed width, skip single-line processing
	if !cf.exceedsWidth {
		itemWidth := lipgloss.Width(singleItem)

		// Add separator width (", ") for non-first items
		if len(cf.singleItems) > 0 {
			itemWidth += 2
		}

		cf.currentWidth += itemWidth

		// Check if we've reached the maximum number of keys to show inline
		if cf.maxKeysInline > 0 && len(cf.singleItems) >= cf.maxKeysInline {
			cf.exceedsWidth = true
		} else {
			// Check if adding this item would exceed the width limit
			closingWidth := lipgloss.Width(cf.closeBrace)
			if cf.currentWidth+closingWidth > cf.p.MaxWidth {
				cf.exceedsWidth = true
			} else {
				cf.singleItems = append(cf.singleItems, singleItem)
			}
		}
	}

	// Always add to multi-line items for fallback
	cf.multiItems = append(cf.multiItems, multiItem)
}

// format returns the final formatted string, choosing single or multi-line based on width
func (cf *compoundFormatter) format() string {
	if len(cf.multiItems) == 0 {
		if cf.typeName != "" {
			return cf.typeName + cf.openBrace + cf.closeBrace
		}
		return cf.openBrace + cf.closeBrace
	}

	// If we exceeded width during processing or don't have all items in single format, use multi-line
	if cf.exceedsWidth || len(cf.singleItems) != len(cf.multiItems) {
		return cf.formatMultiLine()
	}

	// Build single line using pre-calculated width knowledge
	var sb strings.Builder
	if cf.typeName != "" {
		sb.WriteString(cf.typeName)
	}
	sb.WriteString(cf.openBrace)
	if cf.padBraces {
		sb.WriteString(" ")
	}
	for i, item := range cf.singleItems {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(item)
	}
	if cf.padBraces {
		sb.WriteString(" ")
	}
	sb.WriteString(cf.closeBrace)

	return sb.String()
}

// formatMultiLine formats the compound structure in multi-line format
func (cf *compoundFormatter) formatMultiLine() string {
	var sb strings.Builder
	if cf.typeName != "" {
		sb.WriteString(cf.typeName)
	}
	sb.WriteString(cf.openBrace)
	sb.WriteByte('\n')

	indentStr := strings.Repeat("  ", cf.indent+1)
	for i, item := range cf.multiItems {
		if i > 0 {
			sb.WriteString(",\n")
		}
		sb.WriteString(indentStr)
		sb.WriteString(item)
	}

	sb.WriteByte('\n')
	sb.WriteString(strings.Repeat("  ", cf.indent))
	sb.WriteString(cf.closeBrace)

	return sb.String()
}

// isSpecialHandledType checks if a value is a special type that should bypass struct formatting
func (p *Printer) isSpecialHandledType(val reflect.Value) bool {
	return val.Type() == timeType
}

// unwrapInterface extracts the underlying value from interface wrappers
func (p *Printer) unwrapInterface(val reflect.Value) reflect.Value {
	if val.Kind() == reflect.Interface && !val.IsNil() {
		return val.Elem()
	}
	return val
}

// shouldOmitStructName checks if struct name should be omitted based on different criteria
func (p *Printer) shouldOmitStructName(keyOrFieldName string, val reflect.Value, fieldType reflect.Type) bool {
	// Handle interface-wrapped structs
	actualValue := p.unwrapInterface(val)

	// Only consider struct values (including structs pointed to by pointers)
	structVal := actualValue
	if actualValue.Kind() == reflect.Ptr {
		if actualValue.IsNil() || actualValue.Elem().Kind() != reflect.Struct {
			return false
		}
		structVal = actualValue.Elem()
	} else if actualValue.Kind() != reflect.Struct {
		return false
	}

	// If field type is provided, check if it's concrete (not interface)
	if fieldType != nil {
		// If field type is an interface, don't omit struct name
		if fieldType.Kind() == reflect.Interface {
			return false
		}
		// If field type is a pointer to an interface, don't omit struct name
		if fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.Interface {
			return false
		}
		// For concrete types, omit struct name
		return true
	}

	// Otherwise, check if field/key name matches struct type name
	structTypeName := structVal.Type().Name()
	return keyOrFieldName == structTypeName
}

// Print formats any input value into a pretty-printed string representation using default options
func Print(v interface{}) string {
	return Default.Print(v)
}

// formatCyclePointer formats a pointer value for cycle display using Base64 encoding
func (p *Printer) formatCyclePointer(ptr uintptr) string {
	// Hash the pointer to ensure visual distinction between similar pointers
	hasher := fnv.New64a()
	binary.Write(hasher, binary.LittleEndian, uint64(ptr))
	hashedPtr := hasher.Sum64()

	// Convert hashed pointer to byte slice for Base64 encoding
	ptrBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		ptrBytes[i] = byte(hashedPtr >> (8 * i))
	}

	// Encode using standard Base64
	encoded := base64.StdEncoding.EncodeToString(ptrBytes)
	encoded = strings.TrimRight(encoded, "=")

	// Use hash for color selection to maintain consistency
	style := pointerGamut[hashedPtr%uint64(len(pointerGamut))]

	// Format with dim style and parentheses
	return p.colorize("#", p.Styles.Comment) + p.colorize(encoded, style)
}

// isUUID checks if a byte slice represents a valid UUID
func isUUID(data []byte) bool {
	// Standard UUID is 16 bytes
	if len(data) != 16 {
		return false
	}

	// Check UUID version (bits 12-15 of the time_hi_and_version field)
	// Version should be 1-5 for valid UUIDs
	version := (data[6] >> 4) & 0x0f
	if version < 1 || version > 5 {
		return false
	}

	// Check variant (bits 6-7 of the clock_seq_hi_and_reserved field)
	// For RFC 4122 UUIDs, should be 10 binary (0x80-0xBF)
	variant := (data[8] >> 6) & 0x03
	return variant == 2 // 10 binary
}

// formatUUID formats a UUID byte slice using the pointer gamut coloring system
func (p *Printer) formatUUID(data []byte) string {
	// Format as standard UUID string: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	uuidStr := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		data[0:4],
		data[4:6],
		data[6:8],
		data[8:10],
		data[10:16],
	)

	// Use UUID bytes for consistent hash-based color selection
	hasher := fnv.New64a()
	hasher.Write(data)
	hashedUUID := hasher.Sum64()

	// Select color from pointer gamut using same logic as cycle pointers
	style := pointerGamut[hashedUUID%uint64(len(pointerGamut))]

	return p.colorize(uuidStr, style)
}

// isUUIDString checks if a string represents a valid UUID format
func isUUIDString(str string) bool {
	// Standard UUID string format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (36 characters)
	if len(str) != 36 {
		return false
	}

	// Check dash positions
	if str[8] != '-' || str[13] != '-' || str[18] != '-' || str[23] != '-' {
		return false
	}

	// Check that all other characters are valid hex digits
	hexSegments := []string{str[0:8], str[9:13], str[14:18], str[19:23], str[24:36]}
	for _, segment := range hexSegments {
		for _, char := range segment {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
				return false
			}
		}
	}

	return true
}

// formatUUIDString formats a UUID string using the pointer gamut coloring system
func (p *Printer) formatUUIDString(uuidStr string) string {
	// Use the UUID string bytes for consistent hash-based color selection
	hasher := fnv.New64a()
	hasher.Write([]byte(uuidStr))
	hashedUUID := hasher.Sum64()

	// Select color from pointer gamut using same logic as cycle pointers
	style := pointerGamut[hashedUUID%uint64(len(pointerGamut))]

	return p.colorize(uuidStr, style)
}

// tryFormatAsUUID attempts to format a slice or array as a UUID if it contains UUID bytes
func (p *Printer) tryFormatAsUUID(val reflect.Value) string {
	// Only handle byte slices/arrays ([]byte or [N]byte where element type is uint8)
	elemType := val.Type().Elem()
	if elemType.Kind() != reflect.Uint8 {
		return ""
	}

	// Convert to byte slice for UUID checking
	var data []byte
	switch val.Kind() {
	case reflect.Slice:
		if val.IsNil() {
			return ""
		}
		// Convert slice to byte slice
		data = make([]byte, val.Len())
		for i := 0; i < val.Len(); i++ {
			data[i] = byte(val.Index(i).Uint())
		}
	case reflect.Array:
		// Convert array to byte slice
		data = make([]byte, val.Len())
		for i := 0; i < val.Len(); i++ {
			data[i] = byte(val.Index(i).Uint())
		}
	default:
		return ""
	}

	// Check if it's a valid UUID
	if !isUUID(data) {
		return ""
	}

	// Format as UUID with pointer gamut coloring
	return p.formatUUID(data)
}

// appendCyclePointerIfNeeded checks if a value is cycled and appends pointer display
func (p *Printer) appendCyclePointerIfNeeded(formatted string, val reflect.Value) string {
	if !p.canFormCycles(val) {
		return formatted
	}

	var ptr uintptr
	switch val.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice:
		if val.IsNil() {
			return formatted
		}
		ptr = val.Pointer()
	case reflect.Interface:
		// For interfaces, don't append cycle pointers here.
		// The underlying value will handle its own cycle detection and pointer annotation.
		return formatted
	case reflect.Struct:
		// For structs, only track cycle if it's a pointer to a struct
		// Non-pointer structs can't form cycles as they are copied by value
	}

	if ptr != 0 && p.cycled[ptr] {
		return formatted + p.formatCyclePointer(ptr)
	}

	return formatted
}

func PrintWidth(v interface{}, width int) string {
	return New().WithMaxWidth(width).Print(v)
}

// formatValue recursively formats a reflect.Value with proper indentation
func (p *Printer) formatValue(val reflect.Value, indent int) string {
	return p.formatValueWithOptions(val, indent, true)
}

// formatValueWithOptions recursively formats a reflect.Value with formatting options
func (p *Printer) formatValueWithOptions(val reflect.Value, indent int, includeStructNames bool) string {
	if !val.IsValid() {
		return p.colorize("invalid", p.Styles.Error)
	}

	var result string

	// Check for cycles in pointer-like types that can form circular references
	if p.canFormCycles(val) {
		var ptr uintptr
		switch val.Kind() {
		case reflect.Ptr, reflect.Map, reflect.Slice:
			if val.IsNil() {
				// Nil values can't form cycles
				break
			}
			ptr = val.Pointer()
		case reflect.Struct:
			// For structs, only track cycle if it's a pointer to a struct (since
			// that's what can form cycles) Non-pointer structs can't form cycles as
			// they are copied by value. We skip cycle detection for value structs.
			//
			// Note the fields of a value struct may contain pointers to other values
			// that can form cycles, and those are still checked for cycles.
		case reflect.Interface:
			// For interfaces, don't do cycle detection here.
			// Let the underlying value handle its own cycle detection
			// when we recursively call formatValue on val.Elem().
			break
		}

		if ptr != 0 {
			if p.visited[ptr] {
				// Mark this pointer as part of a cycle, but continue with normal formatting
				p.cycled[ptr] = true
				// Return a placeholder for cycled reference
				return p.colorize("→", p.Styles.Comment) + p.formatCyclePointer(ptr)
			}
			// Mark this address as visited
			p.visited[ptr] = true
			// Make sure to clean up after processing this level otherwise we'll prune
			// all further references to this value, despite it not being a cycle.
			// We do NOT clean up the cycled map, because we want to track when a
			// "node" is omitted, and then tag the non-omitted nodes with their ptr.
			defer delete(p.visited, ptr)
		}
	}

	// Check if the value implements io.ReadCloser
	if val.IsValid() && val.CanInterface() {
		if _, ok := val.Interface().(io.ReadCloser); ok {
			result = p.colorize("<io.ReadCloser>", p.Styles.SpecialType)
			return p.appendCyclePointerIfNeeded(result, val)
		}
	}

	if val.Type() == timeType {
		result = p.formatTime(val.Interface().(time.Time))
		return p.appendCyclePointerIfNeeded(result, val)
	}

	switch val.Kind() {
	case reflect.String:
		str := val.String()

		// Check if string is a valid UUID and format it with pointer gamut coloring
		if isUUIDString(str) {
			result = p.formatUUIDString(str)
		} else if js, ok := p.isJSON(str); ok {
			// Check if string is valid JSON and pretty-print it
			if prettyJSON := p.formatJSON(js, indent); prettyJSON != "" {
				result = prettyJSON
			}
		}

		if result == "" {
			// Apply string truncation if needed
			truncatedStr := p.truncateString(str)
			result = p.colorize(fmt.Sprintf(`"%s"`, truncatedStr), p.Styles.String)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		result = p.colorize(fmt.Sprintf("%d", val.Int()), p.Styles.Number)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		result = p.colorize(fmt.Sprintf("%d", val.Uint()), p.Styles.Number)

	case reflect.Float32, reflect.Float64:
		result = p.colorize(fmt.Sprintf("%g", val.Float()), p.Styles.Float)

	case reflect.Bool:
		result = p.colorize(fmt.Sprintf("%t", val.Bool()), p.Styles.Boolean)

	case reflect.Ptr:
		if val.IsNil() {
			result = p.colorize("nil", p.Styles.Null)
		} else {
			result = p.formatValueWithOptions(val.Elem(), indent, includeStructNames)
		}

	case reflect.Interface:
		if val.IsNil() {
			result = p.colorize("nil", p.Styles.Null)
		} else {
			result = p.formatValueWithOptions(val.Elem(), indent, includeStructNames)
		}

	case reflect.Slice, reflect.Array:
		// Check for UUID byte slices first
		if result := p.tryFormatAsUUID(val); result != "" {
			return result
		}
		result = p.formatSlice(val, indent)

	case reflect.Map:
		result = p.formatMap(val, indent)

	case reflect.Struct:
		result = p.formatStruct(val, indent, includeStructNames)

	case reflect.Chan:
		result = p.formatChan(val)

	default:
		// Fallback to JSON for complex types
		if data, err := json.MarshalIndent(val.Interface(), strings.Repeat("  ", indent), "  "); err == nil {
			result = string(data)
		} else {
			result = fmt.Sprintf("%+v", val.Interface())
		}
	}

	return p.appendCyclePointerIfNeeded(result, val)
}

// formatSlice formats slices and arrays with cycle detection
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

	// Use the compound formatter for consistent single/multi-line logic
	formatter := p.newCompoundFormatter("[", "]", "", indent, false, 0)

	for i := 0; i < val.Len(); i++ {
		singleItem := p.formatValue(val.Index(i), 0)       // Single line with 0 indent
		multiItem := p.formatValue(val.Index(i), indent+1) // Multi line with proper indent
		formatter.addItem(singleItem, multiItem)
	}

	return formatter.format()
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
		parts = append(parts, indentStr+p.colorize(truncMsg, p.Styles.Comment))
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
	summary := fmt.Sprintf("// len() = %d", totalLength)
	parts = append(parts, indentStr+p.colorize(summary, p.Styles.Comment))

	return fmt.Sprintf("[\n%s\n%s]", strings.Join(parts, ",\n"), strings.Repeat("  ", indent))
}

// formatMap formats maps with cycle detection
func (p *Printer) formatMap(val reflect.Value, indent int) string {
	if val.Len() == 0 {
		return "{}"
	}

	// Sort keys for consistent output
	keys := val.MapKeys()
	p.sortMapKeys(keys)

	// Use the compound formatter for consistent single/multi-line logic
	formatter := p.newCompoundFormatter("{", "}", "", indent, true, p.MaxKeysInline)

	for _, key := range keys {
		keyStr := p.formatMapKey(key)
		mapValue := val.MapIndex(key)

		// Check if we should omit struct name when key matches struct type
		var singleValueStr, multiValueStr string
		if key.Kind() == reflect.String && !p.isSpecialHandledType(mapValue) {
			// Key matches struct name, format struct without type name
			actualValue := p.unwrapInterface(mapValue)
			omitStructName := p.shouldOmitStructName(key.String(), mapValue, nil)

			// Only call formatStruct if the value is actually a struct
			if omitStructName && actualValue.Kind() == reflect.Struct {
				singleValueStr = p.formatStruct(actualValue, 0, !omitStructName)
				multiValueStr = p.formatStruct(actualValue, indent+1, !omitStructName)
			} else {
				singleValueStr = p.formatValue(mapValue, 0)
				multiValueStr = p.formatValue(mapValue, indent+1)
			}
		} else {
			singleValueStr = p.formatValue(mapValue, 0)
			multiValueStr = p.formatValue(mapValue, indent+1)
		}

		singleItem := fmt.Sprintf("%s: %s", keyStr, singleValueStr)
		multiItem := fmt.Sprintf("%s: %s", keyStr, multiValueStr)
		formatter.addItem(singleItem, multiItem)
	}

	return formatter.format()
}

// formatMapKey formats a map key with cycle detection, treating string keys like struct field names
func (p *Printer) formatMapKey(key reflect.Value) string {
	// If the key is a string, format it like a struct field (no quotes, no coloring)
	if key.Kind() == reflect.String {
		str := key.String()
		// Apply string truncation if needed, but no quotes or styling
		truncatedStr := p.truncateString(str)
		return p.colorize(truncatedStr, p.Styles.Field)
	} else if key.Kind() == reflect.Struct {
		return p.formatStruct(key, 0, false)
	}

	// For non-string keys, use the regular formatting
	return p.formatValue(key, 0)
}

// formatStruct formats structs with optional struct name and cycle detection
func (p *Printer) formatStruct(val reflect.Value, indent int, includeTypeName bool) string {
	typ := val.Type()
	typName := ""
	if includeTypeName {
		typName = typ.Name()
	}
	if val.NumField() == 0 {
		return fmt.Sprintf("%s{}", typName)
	}

	// Use compound formatter
	typeName := ""
	if includeTypeName {
		typeName = typ.Name()
	}
	formatter := p.newCompoundFormatter("{", "}", typeName, indent, true, p.MaxKeysInline)

	// Process exported fields
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldVal := val.Field(i)

		// Check if field has concrete type and omit struct name if so
		var singleFieldStr, multiFieldStr string
		if !p.isSpecialHandledType(fieldVal) && p.shouldOmitStructName(field.Name, fieldVal, field.Type) {
			singleFieldStr = p.formatValueWithOptions(fieldVal, 0, false)
			multiFieldStr = p.formatValueWithOptions(fieldVal, indent+1, false)
		} else {
			singleFieldStr = p.formatValue(fieldVal, 0)
			multiFieldStr = p.formatValue(fieldVal, indent+1)
		}

		singleItem := fmt.Sprintf("%s: %s", field.Name, singleFieldStr)
		multiItem := fmt.Sprintf("%s: %s", field.Name, multiFieldStr)
		formatter.addItem(singleItem, multiItem)
	}

	return formatter.format()
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
func (p *Printer) isJSON(str string) (js json.RawMessage, ok bool) {
	if len(str) < 2 {
		return nil, false
	}

	// Quick check for JSON-like structure
	trimmed := strings.TrimSpace(str)
	if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {

		var js json.RawMessage
		return js, json.Unmarshal([]byte(str), &js) == nil
	}
	return nil, false
}

// formatJSON formats a JSON string with proper indentation and colors
func (p *Printer) formatJSON(jsonStr json.RawMessage, indent int) string {
	var parsed any
	if err := json.Unmarshal(jsonStr, &parsed); err != nil {
		return ""
	}

	// Use our own formatter to format the parsed JSON with colors
	return fmt.Sprintf("%s %s",
		p.colorize("JSON", p.Styles.SpecialType),
		p.formatValue(reflect.ValueOf(parsed), indent))
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

// formatTime formats time.Time values using the relative time formatter
func (p *Printer) formatTime(t time.Time) string {
	// Use the Time function from time.go for humanized relative time
	formatted := Time(t)
	if t.IsZero() {
		// Use special type style for <zero> like other special markers
		return p.colorize(formatted, p.Styles.SpecialType)
	}
	if time.Until(t).Abs() > 30*time.Minute {
		return fmt.Sprintf("%s %s", p.colorize(formatted, p.Styles.Time), p.colorize(t.Format(time.Kitchen), p.Styles.Comment))
	}

	return p.colorize(formatted, p.Styles.Time)
}

// canFormCycles returns true if the given value can potentially form cycles
func (p *Printer) canFormCycles(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice:
		return true
	case reflect.Struct:
		// Structs can form cycles if they contain pointers, slices, or maps
		return true
	case reflect.Interface:
		// Interfaces can contain cycle-forming types
		return true
	default:
		return false
	}
}
