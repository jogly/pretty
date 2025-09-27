package pretty_test

import (
	"fmt"

	"github.com/jogly/pretty"
)

func ExamplePrint_basic() {
	// Basic types
	fmt.Println("String:", pretty.Print("hello world"))
	fmt.Println("Integer:", pretty.Print(42))
	fmt.Println("Boolean:", pretty.Print(true))
	fmt.Println("Nil:", pretty.Print(nil))
	// Output:
	// String: "hello world"
	// Integer: 42
	// Boolean: true
	// Nil: nil
}

func ExamplePrint_collections() {
	// Collections
	slice := []string{"apple", "banana", "cherry"}
	fmt.Println("Slice:")
	fmt.Println(pretty.Print(slice))

	m := map[string]int{"one": 1, "two": 2, "three": 3}
	fmt.Println("Map:")
	fmt.Println(pretty.Print(m))
	// Output:
	// Slice:
	// ["apple", "banana", "cherry"]
	// Map:
	// {one: 1, three: 3, two: 2}
}

func ExamplePrint_struct() {
	// Struct
	type Person struct {
		Name  string
		Age   int
		Email string
	}

	person := Person{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
	}

	pp := &pretty.Printer{MaxWidth: 100}

	fmt.Println("Struct:")
	fmt.Println(pp.Print(person))

	// Output:
	// Struct:
	// Person{Name: "John Doe", Age: 30, Email: "john@example.com"}
}

func ExamplePrint_nested() {
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

	fmt.Println(pretty.Print(person))
	// Output:
	// Person{
	//   Name: "Alice",
	//   Age: 28,
	//   Address: {
	//     Street: "123 Main St",
	//     City: "Anytown",
	//     State: "CA"
	//   },
	//   Hobbies: [
	//     "reading",
	//     "hiking",
	//     "coding"
	//   ]
	// }
}
