//go:build ignore

// Channels
package main

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/jogly/pretty"
)

type Channels struct {
	Bidi chan string
	Send chan<- int
	Recv <-chan bool
}

// START CAPTURE
var pp = pretty.New().WithMaxWidth(40)

func main() {
	channels := Channels{
		Bidi: make(chan string),
		Send: make(chan<- int),
		Recv: make(<-chan bool),
	}

	if len(os.Args) > 1 && os.Args[1] == "spew" {
		spew.Dump(channels)
	} else {
		fmt.Println(pp.Print(channels))
	}
}

// END CAPTURE
