//go:build ignore

package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jogly/pretty"
)

// Demo data structures (reusable across sections)
type UserProfile struct {
	Bio       string
	Website   string
	Skills    []string
	Languages []string
}

type User struct {
	ID       int
	Name     string
	Email    string
	Active   bool
	Created  time.Time
	Profile  UserProfile
	Settings map[string]any
	Tags     []string
}

// Enhanced styling system
var (
	sectionStyle = lipgloss.NewStyle().
			Margin(1, 0)

	sectionTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Bold(true)

	sectionDescStyle = lipgloss.NewStyle().
				Faint(true).
				Italic(true).
				Foreground(lipgloss.Color("244"))

	exampleNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205"))

	exampleDescStyle = lipgloss.NewStyle().
				Faint(true).
				Italic(true).
				Foreground(lipgloss.Color("244"))

	codeStyle = lipgloss.NewStyle().
			Margin(1, 1)

	configStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("211")).
			PaddingBottom(1)
)

func main() {
	// Progressive feature demonstration
	demo1_BasicTypes()
	demo2_CollectionsAndWidth()
	demo3_AdvancedConfigurations()
	demo4_RealWorldStructures()
	demo5_PerformanceFeatures()
	demo6_SpecialCases()
	demo7_CycleDetection()
	demo8_MultipleCycles()
}

func showSection(title, description string) {
	fmt.Println(sectionStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		sectionTextStyle.Render(title),
		sectionDescStyle.Render("// "+description))))
}

func showExample(name, desc, code string) {
	fmt.Println(exampleNameStyle.Render(name))
	if desc != "" {
		fmt.Println(exampleDescStyle.Render("// " + desc))
	}
	fmt.Println(codeStyle.Render(code))
}

func showComparison(name, config1, output1, config2, output2 string) {
	if name != "" {
		fmt.Println(lipgloss.NewStyle().Render(name))
	}

	leftWidth := max(lipgloss.Width(output1), len(name), 15)
	rightWidth := max(lipgloss.Width(output2), len(name), 15)

	leftStyle := lipgloss.NewStyle().Width(leftWidth + 2).Padding(1)
	rightStyle := lipgloss.NewStyle().Width(rightWidth + 2).Padding(1)

	left := lipgloss.JoinVertical(lipgloss.Left, configStyle.Render(config1), output1)
	right := lipgloss.JoinVertical(lipgloss.Left, configStyle.Render(config2), output2)

	joiner := lipgloss.JoinHorizontal
	if lipgloss.Width(left)+lipgloss.Width(right) > 80 {
		joiner = lipgloss.JoinVertical
	}

	fmt.Println(joiner(lipgloss.Top, leftStyle.Render(left), rightStyle.Render(right)))
}

func demo1_BasicTypes() {
	showSection("1. Basic Types & Fundamentals", "basic Go types with colors")

	// Basic types (your examples enhanced)
	showExample("string", "", pretty.Print("hello world"))
	showExample("int", "", pretty.Print(42))
	showExample("float", "", pretty.Print(3.14159))
	showExample("bool", "", pretty.Print(true))
	showExample("nil", "", pretty.Print(nil))

	// Pointers
	str := "hello"
	showExample("*string", "dereferenced automatically", pretty.Print(&str))
}

func demo2_CollectionsAndWidth() {
	showSection("2. Collections & Width Control", "single-line vs multi-line formatting based on width limits")

	// Slices with width comparison (enhanced from your examples)
	slice := []string{"apple", "banana", "cherry", "dragonfruit"}

	narrow := pretty.New().WithMaxWidth(20)
	wide := pretty.New().WithMaxWidth(80)

	showComparison("",
		"MaxWidth: 20", narrow.Print(slice),
		"MaxWidth: 80", wide.Print(slice))

	// Maps with different key types (your examples plus more)
	simpleMap := map[string]int{"one": 1, "two": 2, "three": 3}
	showExample("map[string]int", "keys are sorted", pretty.Print(simpleMap))

	type Key struct{ Name string }
	structMap := map[Key]int{
		{Name: "first"}:  1,
		{Name: "second"}: 2,
		{Name: "third"}:  3,
	}
	showExample("map[Struct]int", "", pretty.PrintWidth(structMap, 30))

	// Complex map
	complexMap := map[string]interface{}{
		"numbers": []int{1, 2, 3, 4, 5},
		"nested":  map[string]string{"inner": "value"},
		"mixed":   "simple string",
	}
	showExample("map[string]any", "", pretty.PrintWidth(complexMap, 30))
}

func demo3_AdvancedConfigurations() {
	showSection("3. Advanced Configurations", "color control, truncation")

	type Person struct {
		Name  string
		Age   int
		Email string
	}
	person := Person{Name: "Alice Johnson", Age: 28, Email: "alice.johnson@company.com"}

	pp := pretty.New().WithMaxWidth(40)
	withColor := pp.WithColorMode(pretty.ColorAlways)
	withColorNever := pp.WithColorMode(pretty.ColorNever)

	// Color modes
	showComparison("",
		"ColorAlways", withColor.Print(person),
		"ColorNever", withColorNever.Print(person))

	// String truncation
	longString := "This is a very long string that demonstrates the truncation feature when MaxStringLength is set"
	showComparison("",
		"60 truncation", pretty.New().WithMaxStringLength(60).Print(longString),
		"30 truncation", pretty.New().WithMaxStringLength(30).Print(longString))
}

func demo4_RealWorldStructures() {
	width := 100
	showSection(
		"4. Real-World Structures",
		fmt.Sprintf("nested structs and complex data with %d characters width", width))

	pp := pretty.New().WithMaxWidth(width)

	user := User{
		ID:      12345,
		Name:    "Sarah Chen",
		Email:   "sarah.chen@techcorp.com",
		Active:  true,
		Created: time.Date(2020, 1, 15, 10, 30, 0, 0, time.UTC),
		Profile: UserProfile{
			Bio:       "Senior software engineer passionate about Go and distributed systems",
			Website:   "https://sarahchen.dev",
			Skills:    []string{"Go", "Kubernetes", "PostgreSQL", "React"},
			Languages: []string{"English", "Mandarin", "Spanish"},
		},
		Settings: map[string]any{
			"theme":         "dark",
			"notifications": true,
			"privacy":       2,
			"features": map[string]bool{
				"beta_features": true,
				"analytics":     false,
			},
		},
		Tags: []string{"engineer", "senior", "backend", "frontend"},
	}

	showExample("User", "", pp.Print(user))

	// API Response simulation
	type APIResponse struct {
		Status    string         `json:"status"`
		Code      int            `json:"code"`
		Message   string         `json:"message"`
		Data      any            `json:"data"`
		Metadata  map[string]any `json:"metadata"`
		Timestamp time.Time      `json:"timestamp"`
	}

	apiResp := APIResponse{
		Status:  "success",
		Code:    200,
		Message: "User retrieved successfully",
		Data:    user,
		Metadata: map[string]any{
			"request_id":    "req-123-abc",
			"response_time": "45ms",
			"cached":        false,
		},
		Timestamp: time.Now(),
	}

	showExample("APIResponse", "", pp.Print(apiResp))
}

func demo5_PerformanceFeatures() {
	showSection("5. Performance Features", "optimizations for large datasets")

	// Slice truncation
	largeSlice := make([]int, 100)
	for i := range largeSlice {
		largeSlice[i] = i + 1
	}

	showComparison("",
		"no limit", pretty.PrintWidth(largeSlice[:10], 20), // Show first 10 for demo
		"MaxSliceLength: 6", pretty.New().WithMaxSliceLength(6).Print(largeSlice))

	// Performance with struct name omission
	type Product struct {
		Name  string
		Price float64
	}

	productMap := map[string]Product{
		"Product": {Name: "Laptop", Price: 999.99}, // Key matches struct name
		"item":    {Name: "Mouse", Price: 29.99},   // Key doesn't match
	}

	showExample("smart struct names", "", pretty.Print(productMap))
}

func demo6_SpecialCases() {
	showSection("6. Special Cases", "channels, interfaces, JSON strings, and edge cases")

	// Channels
	type ChannelDemo struct {
		Bidirectional chan string
		SendOnly      chan<- int
		ReceiveOnly   <-chan bool
	}

	channels := ChannelDemo{
		Bidirectional: make(chan string),
		SendOnly:      make(chan<- int),
		ReceiveOnly:   make(<-chan bool),
	}
	defer close(channels.Bidirectional)
	defer close(channels.SendOnly)

	showExample("channels", "", pretty.Print(channels))

	// io.ReadCloser interface
	type FileDemo struct {
		Reader io.ReadCloser
		Name   string
		Size   int64
	}

	fileDemo := FileDemo{
		Reader: io.NopCloser(strings.NewReader("file contents")),
		Name:   "example.txt",
		Size:   1024,
	}

	showExample("io.ReadCloser", "", pretty.Print(fileDemo))

	// JSON string formatting
	jsonData := struct {
		RawJSON    string
		ParsedData string
		Invalid    string
	}{
		RawJSON:    `{"name":"John","age":30,"skills":["Go","Rust"],"active":true}`,
		ParsedData: "regular string",
		Invalid:    `{invalid:"json"}`,
	}

	showExample("JSON strings", "auto-detects and formats JSON", pretty.Print(jsonData))

	// Deep nesting showcase
	type Node struct {
		Value    string
		Children []Node
		Metadata map[string]string
	}

	deepStruct := Node{
		Value: "Root",
		Children: []Node{
			{
				Value: "Child1",
				Children: []Node{
					{Value: "Grandchild1", Metadata: map[string]string{"type": "leaf"}},
					{Value: "Grandchild2", Metadata: map[string]string{"type": "leaf"}},
				},
				Metadata: map[string]string{"level": "2"},
			},
			{Value: "Child2", Metadata: map[string]string{"level": "2"}},
		},
		Metadata: map[string]string{"level": "1", "root": "true"},
	}

	showExample("deep nesting", "", pretty.New().WithMaxWidth(80).Print(deepStruct))

	// Edge cases
	var nilStruct *User
	emptyStruct := User{}

	showComparison("",
		"nil struct", pretty.Print(nilStruct),
		"empty struct", pretty.Print(emptyStruct))
}

func demo7_CycleDetection() {
	showSection("7. Cycle Detection", "prevents infinite recursion in circular references")

	// Self-referencing struct
	type CircularNode struct {
		ID    int
		Value string
		Next  *CircularNode
	}

	node1 := &CircularNode{ID: 1, Value: "first"}
	node2 := &CircularNode{ID: 2, Value: "second", Next: node1}
	node1.Next = node2 // Create cycle

	showExample("self-referencing struct", "", pretty.Print(node1))

	// Circular slice
	circularSlice := make([]interface{}, 2)
	circularSlice[0] = "item"
	circularSlice[1] = circularSlice // Self-reference

	showExample("circular slice", "", pretty.Print(circularSlice))

	// Circular map
	circularMap := make(map[string]interface{})
	circularMap["key"] = "value"
	circularMap["self"] = circularMap // Self-reference

	showExample("circular map", "", pretty.Print(circularMap))

	// Indirect circular reference
	type Container struct {
		Name string
		Ref  interface{}
	}

	container1 := &Container{Name: "first"}
	container2 := &Container{Name: "second", Ref: container1}
	container1.Ref = container2 // Create indirect cycle

	showExample("indirect cycle via interface", "", pretty.Print(container1))
}

func demo8_MultipleCycles() {
	showSection("8. Multiple Cycles", "multiple cycles in a single structure")

	// Multiple cycles
	type MultipleCycleNode struct {
		ID   int
		Next *MultipleCycleNode
		Tail *MultipleCycleNode
		Prev *MultipleCycleNode
	}

	node1 := &MultipleCycleNode{ID: 1}
	node2 := &MultipleCycleNode{ID: 2, Next: node1}
	node3 := &MultipleCycleNode{ID: 3, Next: node2}
	node1.Next = node3

	node1.Tail = node3
	node2.Tail = node3
	node3.Prev = node2
	node3.Tail = node3

	showExample("multiple cycles", "", pretty.Print(node1))
}
