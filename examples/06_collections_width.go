//go:build ignore

// Slices
package main

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/jogly/pretty"
)

// START CAPTURE
var pp = pretty.New().WithMargin(0, 0, 1, 0)

func main() {

	fruit := []string{"apple", "banana", "cherry", "dragonfruit"}

	if len(os.Args) > 1 && os.Args[1] == "spew" {
		spew.Dump(fruit)
	} else {
		fmt.Println(pp.Print(fruit))
		pp = pp.WithMaxWidth(20)
		fmt.Println(pp.Print(fruit))
	}

}

// END CAPTURE
