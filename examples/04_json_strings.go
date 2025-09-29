//go:build ignore

// auto-detects and formats JSON
package main

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/jogly/pretty"
)

// START CAPTURE
var pp = pretty.New().WithMaxWidth(40)

func main() {
	data := struct {
		RawJSON string
		NotJSON string
	}{
		RawJSON: `{"name":"John","age":30,"skills":["Go","Rust"],"active":true}`,
		NotJSON: `{invalid:"json"}`,
	}

	if len(os.Args) > 1 && os.Args[1] == "spew" {
		spew.Dump(data)
	} else {
		fmt.Println(pp.Print(data))
	}
}

// END CAPTURE
