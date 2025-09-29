//go:build ignore

// References with colorful anchors
package main

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/jogly/pretty"
)

type Node struct {
	ID    int
	Value string
	Next  *Node
}

// START CAPTURE
var pp = pretty.New().WithMaxWidth(40)

func main() {
	// Self-referencing struct
	node1 := &Node{ID: 1, Value: "first"}
	node2 := &Node{ID: 2, Value: "second", Next: node1}
	node1.Next = node2 // Create cycle

	if len(os.Args) > 1 && os.Args[1] == "spew" {
		spew.Dump(node1)
	} else {
		fmt.Println(pp.Print(node1))
	}
}

// END CAPTURE
