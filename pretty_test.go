package pretty

import (
	"io"
	"strings"
	"testing"
	"time"
)

type TestStruct struct {
	Name  string
	Age   int
	Email string
}

type NestedStruct struct {
	User   TestStruct
	Active bool
	Tags   []string
}

type Point struct {
	X int
	Y int
}

func TestPrint(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: "nil",
		},
		{
			name:     "string",
			input:    "hello world",
			expected: `"hello world"`,
		},
		{
			name:     "integer",
			input:    42,
			expected: "42",
		},
		{
			name:     "float",
			input:    3.14159,
			expected: "3.14159",
		},
		{
			name:     "boolean true",
			input:    true,
			expected: "true",
		},
		{
			name:     "boolean false",
			input:    false,
			expected: "false",
		},
		{
			name:     "empty slice",
			input:    []int{},
			expected: "[]",
		},
		{
			name:     "slice with elements",
			input:    []int{1, 2, 3},
			expected: "[1, 2, 3]",
		},
		{
			name:     "empty map",
			input:    map[string]int{},
			expected: "{}",
		},
		{
			name:     "map with elements",
			input:    map[string]int{"a": 1, "b": 2},
			expected: "{a: 1, b: 2}",
		},
		{
			name:     "simple struct",
			input:    TestStruct{Name: "John", Age: 30, Email: "john@example.com"},
			expected: "TestStruct{\n  Name: \"John\",\n  Age: 30,\n  Email: \"john@example.com\"\n}",
		},
		{
			name: "nested struct",
			input: NestedStruct{
				User:   TestStruct{Name: "Alice", Age: 25, Email: "alice@example.com"},
				Active: true,
				Tags:   []string{"admin", "user"},
			},
			expected: "NestedStruct{\n  User: TestStruct{\n    Name: \"Alice\",\n    Age: 25,\n    Email: \"alice@example.com\"\n  },\n  Active: true,\n  Tags: [\"admin\", \"user\"]\n}",
		},
		{
			name:     "long slice multi-line",
			input:    []string{"very_long_string_name", "another_very_long_string", "third_long_string"},
			expected: "[\n  \"very_long_string_name\",\n  \"another_very_long_string\",\n  \"third_long_string\"\n]",
		},
		{
			name:     "long map multi-line",
			input:    map[string]string{"very_long_key": "very_long_value", "another_key": "another_value"},
			expected: "{\n  another_key: \"another_value\",\n  very_long_key: \"very_long_value\"\n}",
		},
		{
			name:     "short struct single-line",
			input:    Point{X: 10, Y: 20},
			expected: "Point{X: 10, Y: 20}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Print(tt.input)
			if result != tt.expected {
				t.Errorf("Print() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPrintPointers(t *testing.T) {
	str := "hello"
	var nilPtr *string

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "pointer to string",
			input:    &str,
			expected: `"hello"`,
		},
		{
			name:     "nil pointer",
			input:    nilPtr,
			expected: "nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Print(tt.input)
			if result != tt.expected {
				t.Errorf("Print() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPrintChannels(t *testing.T) {
	type ChannelStruct struct {
		Both chan int
		Send chan<- int
		Recv <-chan int
	}

	ch := ChannelStruct{
		Both: make(chan int),
		Send: make(chan<- int),
		Recv: make(<-chan int),
	}

	// Close channels to avoid blocking
	close(ch.Both)
	close(ch.Send)

	result := Print(ch)
	expected := "ChannelStruct{\n  Both: chan int,\n  Send: chan<- int,\n  Recv: <-chan int\n}"
	if result != expected {
		t.Errorf("Print(channels) = %q, want %q", result, expected)
	}

	// Test individual channel types
	var sendOnly chan<- string = make(chan<- string)
	var recvOnly <-chan string = make(<-chan string)
	var biDir chan string = make(chan string)

	close(sendOnly)
	close(biDir)

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"send-only", sendOnly, "chan<- string"},
		{"recv-only", recvOnly, "<-chan string"},
		{"bidirectional", biDir, "chan string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Print(tt.input)
			if result != tt.expected {
				t.Errorf("Print(%s) = %q, want %q", tt.name, result, tt.expected)
			}
		})
	}
}

func TestPrintComplexTypes(t *testing.T) {
	// Test time.Time which has unexported fields
	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	result := Print(now)

	// Should show as Time{} since it has unexported fields
	expected := "Time{}"
	if result != expected {
		t.Errorf("Print(time.Time) = %q, want %q", result, expected)
	}
}

func TestPrinterWithCustomWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		width    int
		expected string
	}{
		{
			name:     "slice with narrow width",
			input:    []int{1, 2, 3, 4},
			width:    10,
			expected: "[\n  1,\n  2,\n  3,\n  4\n]",
		},
		{
			name:     "slice with wide width",
			input:    []int{1, 2, 3, 4},
			width:    50,
			expected: "[1, 2, 3, 4]",
		},
		{
			name:     "map with narrow width",
			input:    map[string]int{"a": 1, "b": 2},
			width:    10,
			expected: "{\n  a: 1,\n  b: 2\n}",
		},
		{
			name:     "map with wide width",
			input:    map[string]int{"a": 1, "b": 2},
			width:    50,
			expected: "{a: 1, b: 2}",
		},
		{
			name:     "struct with narrow width",
			input:    Point{X: 10, Y: 20},
			width:    15,
			expected: "Point{\n  X: 10,\n  Y: 20\n}",
		},
		{
			name:     "struct with wide width",
			input:    Point{X: 10, Y: 20},
			width:    50,
			expected: "Point{X: 10, Y: 20}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printer := &Printer{MaxWidth: tt.width}
			result := printer.Print(tt.input)
			if result != tt.expected {
				t.Errorf("Printer.Print() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNew(t *testing.T) {
	printer := New()
	if printer.MaxWidth != 30 {
		t.Errorf("New().MaxWidth = %d, want 30", printer.MaxWidth)
	}
}

func TestWithMaxWidth(t *testing.T) {
	printer := New().WithMaxWidth(50)
	if printer.MaxWidth != 50 {
		t.Errorf("WithMaxWidth(50).MaxWidth = %d, want 50", printer.MaxWidth)
	}

	// Test chaining
	data := []int{1, 2, 3, 4, 5}
	result := New().WithMaxWidth(100).Print(data)
	expected := "[1, 2, 3, 4, 5]"
	if result != expected {
		t.Errorf("Chained WithMaxWidth().Print() = %q, want %q", result, expected)
	}
}

func TestMapKeySorting(t *testing.T) {
	// Test that map keys are sorted alphabetically
	m := map[string]int{
		"zebra": 1,
		"alpha": 2,
		"beta":  3,
		"gamma": 4,
	}

	// Use a wide printer to ensure single-line output
	printer := &Printer{MaxWidth: 100}
	result := printer.Print(m)
	expected := "{alpha: 2, beta: 3, gamma: 4, zebra: 1}"
	if result != expected {
		t.Errorf("Map keys not sorted correctly: got %q, want %q", result, expected)
	}

	// Test with multi-line format
	longMap := map[string]string{
		"zzz_very_long_key": "value1",
		"aaa_another_key":   "value2",
		"mmm_middle_key":    "value3",
	}

	narrowPrinter := &Printer{MaxWidth: 10} // Force multi-line
	result = narrowPrinter.Print(longMap)
	expected = "{\n  aaa_another_key: \"value2\",\n  mmm_middle_key: \"value3\",\n  zzz_very_long_key: \"value1\"\n}"
	if result != expected {
		t.Errorf("Multi-line map keys not sorted correctly:\ngot:  %q\nwant: %q", result, expected)
	}
}

// mockReadCloser implements io.ReadCloser for testing
type mockReadCloser struct {
	*strings.Reader
}

func (m *mockReadCloser) Close() error {
	return nil
}

func newMockReadCloser(content string) io.ReadCloser {
	return &mockReadCloser{
		Reader: strings.NewReader(content),
	}
}

func TestReadCloserFormatting(t *testing.T) {
	// Test with mockReadCloser (implements ReadCloser)
	readCloser := newMockReadCloser("test content")
	defer readCloser.Close()

	printer := New().WithColorMode(ColorNever)
	result := printer.Print(readCloser)
	expected := "<io.ReadCloser>"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	// Test with strings.Reader (does not implement ReadCloser)
	reader := strings.NewReader("test content")
	result = printer.Print(reader)

	// Should not show <io.ReadCloser> since strings.Reader doesn't implement io.ReadCloser
	if result == "<io.ReadCloser>" {
		t.Error("strings.Reader should not be formatted as <io.ReadCloser>")
	}

	// Test in a struct
	data := struct {
		ReadCloser io.ReadCloser
		Reader     *strings.Reader
	}{
		ReadCloser: readCloser,
		Reader:     reader,
	}

	result = printer.Print(data)
	if !strings.Contains(result, "<io.ReadCloser>") {
		t.Error("Expected struct to contain <io.ReadCloser> for ReadCloser field")
	}
}

func TestJSONFormatting(t *testing.T) {
	printer := New().WithColorMode(ColorNever)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple JSON object",
			input:    `{"name":"John","age":30}`,
			expected: "{age: 30, name: \"John\"}",
		},
		{
			name:     "JSON array",
			input:    `[1,2,3]`,
			expected: "[1, 2, 3]",
		},
		{
			name:     "nested JSON",
			input:    `{"user":{"name":"Alice","active":true},"count":5}`,
			expected: "{\n  count: 5,\n  user: {active: true, name: \"Alice\"}\n}",
		},
		{
			name:     "not JSON - regular string",
			input:    "hello world",
			expected: `"hello world"`,
		},
		{
			name:     "invalid JSON - looks like JSON but isn't",
			input:    `{name:"John"}`,
			expected: `"{name:"John"}"`,
		},
		{
			name:     "empty JSON object",
			input:    `{}`,
			expected: "{}",
		},
		{
			name:     "empty JSON array",
			input:    `[]`,
			expected: "[]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := printer.Print(tt.input)
			if result != tt.expected {
				t.Errorf("Print(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestJSONInStructs(t *testing.T) {
	printer := New().WithColorMode(ColorNever).WithMaxWidth(100)

	data := struct {
		Config string
		Data   string
	}{
		Config: `{"debug":true,"timeout":30}`,
		Data:   "regular string",
	}

	result := printer.Print(data)

	// Should contain pretty-printed JSON for Config field
	if !strings.Contains(result, `debug: true`) {
		t.Error("Expected Config field to contain pretty-printed JSON")
	}

	// Should contain quoted string for Data field
	if !strings.Contains(result, `"regular string"`) {
		t.Error("Expected Data field to contain quoted regular string")
	}
}

func TestSliceTruncation(t *testing.T) {
	printer := New().WithColorMode(ColorNever).WithMaxSliceLength(6)

	tests := []struct {
		name     string
		input    []int
		expected string
	}{
		{
			name:     "short slice - no truncation",
			input:    []int{1, 2, 3},
			expected: "[1, 2, 3]",
		},
		{
			name:     "exact limit - no truncation",
			input:    []int{1, 2, 3, 4, 5, 6},
			expected: "[1, 2, 3, 4, 5, 6]",
		},
		{
			name:     "long slice - truncated",
			input:    []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			expected: "[\n  1,\n  2,\n  3,\n  ... 4 more elements ...,\n  8,\n  9,\n  10,\n  // Total length: 10\n]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := printer.Print(tt.input)
			if result != tt.expected {
				t.Errorf("Print(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSliceTruncationDisabled(t *testing.T) {
	// Test that MaxSliceLength = 0 disables truncation
	printer := New().WithColorMode(ColorNever).WithMaxSliceLength(0)

	longSlice := make([]int, 100)
	for i := 0; i < 100; i++ {
		longSlice[i] = i + 1
	}

	result := printer.Print(longSlice)

	// Should contain all elements, not truncated
	if !strings.Contains(result, "100") {
		t.Error("Expected full slice to contain last element 100")
	}
	if strings.Contains(result, "more elements") {
		t.Error("Expected no truncation when MaxSliceLength = 0")
	}
}

func TestWithMaxSliceLength(t *testing.T) {
	printer := New().WithMaxSliceLength(4)
	if printer.MaxSliceLength != 4 {
		t.Errorf("WithMaxSliceLength(4).MaxSliceLength = %d, want 4", printer.MaxSliceLength)
	}

	// Test chaining
	result := New().WithMaxSliceLength(2).WithColorMode(ColorNever)
	if result.MaxSliceLength != 2 {
		t.Errorf("Chained WithMaxSliceLength(2).MaxSliceLength = %d, want 2", result.MaxSliceLength)
	}
	if result.ColorMode != ColorNever {
		t.Errorf("Chained WithMaxSliceLength(2).WithColorMode(ColorNever).ColorMode = %v, want ColorNever", result.ColorMode)
	}
}

func TestStringTruncation(t *testing.T) {
	printer := New().WithColorMode(ColorNever).WithMaxStringLength(20)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short string - no truncation",
			input:    "hello world",
			expected: `"hello world"`,
		},
		{
			name:     "exact limit - no truncation",
			input:    "twelve characters!!",
			expected: `"twelve characters!!"`,
		},
		{
			name:     "long string - truncated",
			input:    "this is a very long string that should be truncated",
			expected: `"this is ...truncated"`,
		},
		{
			name:     "medium truncation",
			input:    "hello world from testing",
			expected: `"hello wo...m testing"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := printer.Print(tt.input)
			if result != tt.expected {
				t.Errorf("Print(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStringTruncationEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		maxLen   int
		input    string
		expected string
	}{
		{
			name:     "very small limit - fallback",
			maxLen:   3,
			input:    "hello world",
			expected: `"hel"`,
		},
		{
			name:     "minimum truncation size",
			maxLen:   5,
			input:    "hello world",
			expected: `"h...d"`,
		},
		{
			name:     "disabled truncation",
			maxLen:   0,
			input:    "this is a very long string that should not be truncated",
			expected: `"this is a very long string that should not be truncated"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printer := New().WithColorMode(ColorNever).WithMaxStringLength(tt.maxLen)
			result := printer.Print(tt.input)
			if result != tt.expected {
				t.Errorf("Print(%q) with maxLen=%d = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestStringTruncationInStructs(t *testing.T) {
	printer := New().WithColorMode(ColorNever).WithMaxStringLength(15).WithMaxWidth(100)

	data := struct {
		Short string
		Long  string
	}{
		Short: "hello",
		Long:  "this is a very long field that should be truncated",
	}

	result := printer.Print(data)

	// Should contain untruncated short string
	if !strings.Contains(result, `"hello"`) {
		t.Error("Expected Short field to remain untruncated")
	}

	// Should contain truncated long string - debug what we actually get
	if !strings.Contains(result, "...") {
		t.Errorf("Expected Long field to be truncated, got: %s", result)
	}
}

func TestWithMaxStringLength(t *testing.T) {
	printer := New().WithMaxStringLength(25)
	if printer.MaxStringLength != 25 {
		t.Errorf("WithMaxStringLength(25).MaxStringLength = %d, want 25", printer.MaxStringLength)
	}

	// Test chaining
	result := New().WithMaxStringLength(10).WithColorMode(ColorNever)
	if result.MaxStringLength != 10 {
		t.Errorf("Chained WithMaxStringLength(10).MaxStringLength = %d, want 10", result.MaxStringLength)
	}
	if result.ColorMode != ColorNever {
		t.Errorf("Chained WithMaxStringLength(10).WithColorMode(ColorNever).ColorMode = %v, want ColorNever", result.ColorMode)
	}
}

func TestStructNameOmissionInMaps(t *testing.T) {
	// Define test structs
	type Person struct {
		Name string
		Age  int
	}

	type Address struct {
		Street string
		City   string
	}

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name: "map with string key matching struct name",
			input: map[string]Person{
				"Person": {Name: "John", Age: 30},
			},
			expected: `{Person: {Name: "John", Age: 30}}`,
		},
		{
			name: "map with string key not matching struct name",
			input: map[string]Person{
				"user": {Name: "John", Age: 30},
			},
			expected: `{user: Person{Name: "John", Age: 30}}`,
		},
		{
			name: "map with multiple entries, some matching struct names",
			input: map[string]interface{}{
				"Person":  Person{Name: "John", Age: 30},
				"Address": Address{Street: "123 Main St", City: "Anytown"},
				"other":   Person{Name: "Jane", Age: 25},
			},
			expected: `{Address: {Street: "123 Main St", City: "Anytown"}, Person: {Name: "John", Age: 30}, other: Person{Name: "Jane", Age: 25}}`,
		},
		{
			name: "map with non-string keys should not affect struct names",
			input: map[int]Person{
				1: {Name: "John", Age: 30},
			},
			expected: `{1: Person{Name: "John", Age: 30}}`,
		},
		{
			name: "map with non-struct values should not be affected",
			input: map[string]string{
				"Person": "not a struct",
			},
			expected: `{Person: "not a struct"}`,
		},
		{
			name: "struct field name matches struct type name",
			input: struct {
				Person struct {
					Name string
					Age  int
				}
				Other string
			}{
				Person: struct {
					Name string
					Age  int
				}{Name: "John", Age: 30},
				Other: "test",
			},
			expected: `{Person: {Name: "John", Age: 30}, Other: "test"}`,
		},
	}

	// Use ColorNever to avoid color codes in test comparisons
	printer := New().WithColorMode(ColorNever).WithMaxWidth(200)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := printer.Print(tt.input)
			if result != tt.expected {
				t.Errorf("Print() = %q, want %q", result, tt.expected)
			}
		})
	}
}
