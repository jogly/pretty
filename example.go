//go:build ignore

package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/jogly/pretty"
)

func main() {
	ExamplePrint_basic()
	ExamplePrint_collections()
	ExamplePrint_struct()
	ExamplePrint_nested()
}

func showExample(name, formatted string) {
	desc := lipgloss.NewStyle().
		Faint(true).
		Bold(true)
	view := lipgloss.NewStyle().Padding(1, 2)
	fmt.Println(desc.Render("// " + name))
	fmt.Println(view.Render(formatted))
}

func ExamplePrint_basic() {
	// Basic types
	showExample("String", pretty.Print("hello world"))
	showExample("Integer", pretty.Print(42))
	showExample("Boolean", pretty.Print(true))
	showExample("Nil", pretty.Print(nil))
}

func ExamplePrint_collections() {
	// Collections
	pp := pretty.New().WithMaxWidth(30)
	slice := []string{"apple", "banana", "cherry"}
	showExample("Slice (30 chars)", pp.Print(slice))
	pp = pp.WithMaxWidth(80)
	showExample("Slice (80 chars)", pp.Print(slice))

	m := map[string]int{"one": 1, "two": 2, "three": 3}
	showExample("Small map", pp.Print(m))

	type key struct{ Key string }
	m2 := map[key]int{
		{Key: "one"}:   1,
		{Key: "two"}:   2,
		{Key: "three"}: 3,
	}
	showExample("Map with struct key", pretty.Print(m2))
}

func ExamplePrint_struct() {
	// Struct
	type Person struct {
		Name  string
		Age   int
		Email string
	}

	person := Person{Name: "John Doe", Age: 30, Email: "john@example.com"}
	showExample("Person", pretty.Print(person))
}

func ExamplePrint_nested() {
	// Nested
	type Address struct {
		Street string
		City   string
		State  string
	}

	type Person struct {
		Name    string
		Age     int
		Address Address
		Hobbies []string
	}

	person := Person{
		Name: "Alice",
		Age:  28,
		Address: Address{
			Street: "123 Main St",
			City:   "Anytown",
			State:  "CA",
		},
		Hobbies: []string{"reading", "hiking", "coding"},
	}

	showExample("Nested", pretty.Print(person))
}
