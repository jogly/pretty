//go:build ignore

// Large slices
package main

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/jogly/pretty"
)

type Product struct {
	Name  string
	Price float64
}

// START CAPTURE
var pp = pretty.New().
	WithMaxWidth(40).
	WithMaxSliceLength(6)

func main() {
	// Large slice truncation
	largeSlice := make([]int, 100)
	for i := range largeSlice {
		largeSlice[i] = i + 1
	}

	if len(os.Args) > 1 && os.Args[1] == "spew" {
		spew.Dump(largeSlice)
	} else {
		fmt.Println(pp.Print(largeSlice))
	}
}

// END CAPTURE
