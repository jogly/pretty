package pretty

import (
	"io"
	"strings"
	"testing"
	"time"
)

// Complex nested data structures for benchmarking
type ComplexUser struct {
	ID       int
	Name     string
	Email    string
	Age      int
	Active   bool
	Tags     []string
	Profile  UserProfile
	Settings map[string]interface{}
	Created  time.Time
	Reader   io.ReadCloser
}

type UserProfile struct {
	Bio        string
	Website    string
	Location   string
	Languages  []string
	Skills     []Skill
	Experience []Experience
}

type Skill struct {
	Name       string
	Level      int
	Certified  bool
	Categories []string
}

type Experience struct {
	Company     string
	Title       string
	Duration    string
	Description string
	Skills      []string
}

// Create complex test data
func createComplexUser() ComplexUser {
	return ComplexUser{
		ID:      12345,
		Name:    "John Doe Anderson Smith Jr.",
		Email:   "john.doe.anderson.smith.jr@verylongcompanyname.com",
		Age:     35,
		Active:  true,
		Tags:    []string{"admin", "developer", "team-lead", "architect", "mentor"},
		Created: time.Date(2020, 1, 15, 10, 30, 0, 0, time.UTC),
		Reader:  &mockReadCloser{strings.NewReader("sample data")},
		Profile: UserProfile{
			Bio:      "Experienced software engineer with 15+ years in full-stack development, cloud architecture, and team leadership. Passionate about mentoring junior developers and building scalable systems.",
			Website:  "https://johndoe.dev",
			Location: "San Francisco, CA, USA",
			Languages: []string{"English", "Spanish", "French", "German"},
			Skills: []Skill{
				{
					Name:       "Go Programming",
					Level:      9,
					Certified:  true,
					Categories: []string{"backend", "systems", "cloud"},
				},
				{
					Name:       "JavaScript",
					Level:      8,
					Certified:  false,
					Categories: []string{"frontend", "backend", "mobile"},
				},
				{
					Name:       "Docker & Kubernetes",
					Level:      7,
					Certified:  true,
					Categories: []string{"devops", "cloud", "containers"},
				},
			},
			Experience: []Experience{
				{
					Company:     "TechCorp International Inc.",
					Title:       "Senior Software Engineering Manager",
					Duration:    "2020-present",
					Description: "Leading a team of 12 engineers working on microservices architecture using Go, Docker, and Kubernetes in AWS cloud environment.",
					Skills:      []string{"Go", "Docker", "Kubernetes", "AWS", "Team Leadership"},
				},
				{
					Company:     "StartupXYZ Ltd.",
					Title:       "Full Stack Developer",
					Duration:    "2017-2020",
					Description: "Built scalable web applications using React, Node.js, and PostgreSQL. Implemented CI/CD pipelines and automated testing.",
					Skills:      []string{"React", "Node.js", "PostgreSQL", "CI/CD"},
				},
			},
		},
		Settings: map[string]interface{}{
			"theme":           "dark",
			"notifications":   true,
			"language":        "en-US",
			"timezone":        "America/Los_Angeles",
			"privacy_level":   2,
			"email_frequency": "weekly",
			"features": map[string]bool{
				"beta_features":     true,
				"analytics":         false,
				"marketing_emails":  false,
				"social_sharing":    true,
				"two_factor_auth":   true,
				"password_recovery": true,
			},
			"preferences": map[string]interface{}{
				"color_scheme":    "blue",
				"font_size":       14,
				"line_height":     1.5,
				"sidebar_collapsed": false,
				"auto_save":       true,
				"recent_files":    []string{"/home/user/project1.go", "/home/user/project2.js", "/home/user/config.yaml"},
			},
		},
	}
}

// Create large slice with complex elements
func createLargeComplexSlice(size int) []ComplexUser {
	users := make([]ComplexUser, size)
	for i := 0; i < size; i++ {
		user := createComplexUser()
		user.ID = i + 1
		user.Name = user.Name + " #" + string(rune(i%26+'A'))
		users[i] = user
	}
	return users
}

// Create large map with complex values
func createLargeComplexMap(size int) map[string]ComplexUser {
	users := make(map[string]ComplexUser, size)
	for i := 0; i < size; i++ {
		user := createComplexUser()
		user.ID = i + 1
		key := "user_" + string(rune(i%26+'a')) + "_" + string(rune((i/26)%26+'a'))
		users[key] = user
	}
	return users
}

// Benchmark single complex struct
func BenchmarkPrint_ComplexStruct(b *testing.B) {
	user := createComplexUser()
	printer := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = printer.Print(user)
	}
}

// Benchmark large slice of complex structs
func BenchmarkPrint_LargeSlice(b *testing.B) {
	users := createLargeComplexSlice(100)
	printer := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = printer.Print(users)
	}
}

// Benchmark large map of complex structs
func BenchmarkPrint_LargeMap(b *testing.B) {
	users := createLargeComplexMap(50)
	printer := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = printer.Print(users)
	}
}

// Benchmark deeply nested structures
func BenchmarkPrint_DeeplyNested(b *testing.B) {
	// Create a deeply nested structure
	type Node struct {
		Value    string
		Children []Node
		Metadata map[string]interface{}
	}

	var createNode func(depth int) Node
	createNode = func(depth int) Node {
		node := Node{
			Value: "Node at depth " + string(rune(depth+'0')),
			Metadata: map[string]interface{}{
				"depth":     depth,
				"timestamp": time.Now(),
				"active":    depth%2 == 0,
				"tags":      []string{"tag1", "tag2", "tag3"},
			},
		}

		if depth > 0 {
			node.Children = []Node{
				createNode(depth - 1),
				createNode(depth - 1),
			}
		}

		return node
	}

	deepNode := createNode(6) // Creates 2^6-1 = 63 total nodes
	printer := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = printer.Print(deepNode)
	}
}

// Benchmark JSON formatting
func BenchmarkPrint_JSONStrings(b *testing.B) {
	data := struct {
		SimpleJSON   string
		ComplexJSON  string
		NestedJSON   string
		InvalidJSON  string
		LargeJSON    string
	}{
		SimpleJSON:  `{"name":"John","age":30,"active":true}`,
		ComplexJSON: `{"users":[{"id":1,"name":"Alice","profile":{"bio":"Developer","skills":["Go","JS"]}}],"meta":{"count":1,"version":"1.0"}}`,
		NestedJSON:  `{"level1":{"level2":{"level3":{"level4":{"value":"deep"}}}}}`,
		InvalidJSON: `{name:"John",age:30}`, // Invalid JSON
		LargeJSON:   `{"data":[` + strings.Repeat(`{"field1":"value1","field2":"value2","field3":"value3"},`, 50)[:49*50] + `]}`,
	}

	printer := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = printer.Print(data)
	}
}

// Benchmark with different width settings
func BenchmarkPrint_DifferentWidths(b *testing.B) {
	user := createComplexUser()

	widths := []int{20, 50, 80, 120, 200}

	for _, width := range widths {
		b.Run("width_"+string(rune(width/10+'0'))+string(rune((width%10)+'0')), func(b *testing.B) {
			printer := New().WithMaxWidth(width)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = printer.Print(user)
			}
		})
	}
}

// Benchmark slice truncation
func BenchmarkPrint_SliceTruncation(b *testing.B) {
	largeSlice := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		largeSlice[i] = i
	}

	b.Run("no_truncation", func(b *testing.B) {
		printer := New().WithMaxSliceLength(0) // No truncation

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = printer.Print(largeSlice)
		}
	})

	b.Run("with_truncation", func(b *testing.B) {
		printer := New().WithMaxSliceLength(20) // Truncate at 20

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = printer.Print(largeSlice)
		}
	})
}

// Benchmark string truncation
func BenchmarkPrint_StringTruncation(b *testing.B) {
	data := struct {
		Short  string
		Medium string
		Long   string
	}{
		Short:  "short",
		Medium: "this is a medium length string that might be truncated",
		Long:   strings.Repeat("this is a very long string that will definitely be truncated ", 10),
	}

	b.Run("no_truncation", func(b *testing.B) {
		printer := New().WithMaxStringLength(0) // No truncation

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = printer.Print(data)
		}
	})

	b.Run("with_truncation", func(b *testing.B) {
		printer := New().WithMaxStringLength(50) // Truncate at 50

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = printer.Print(data)
		}
	})
}

// Benchmark different printer creation methods
func BenchmarkPrinter_Creation(b *testing.B) {
	b.Run("new_printer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = New()
		}
	})

	b.Run("with_chain", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = New().WithMaxWidth(50).WithColorMode(ColorNever).WithMaxSliceLength(100).WithMaxStringLength(80)
		}
	})
}

// Memory allocation benchmark
func BenchmarkPrint_Allocations(b *testing.B) {
	user := createComplexUser()
	printer := New()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result := printer.Print(user)
		_ = result // Prevent optimization
	}
}