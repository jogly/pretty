//go:build ignore

// UUID detection for both byte slices and strings
package main

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/jogly/pretty"
)

type UUIDDemo struct {
	ByteUUID   []byte
	StringUUID string
}

// START CAPTURE
var pp = pretty.New().WithMaxWidth(40)

func main() {
	validUUID := []byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x41, 0x1f, 0xad, 0xc8, 0x00, 0x0c, 0x29, 0x48, 0xe9, 0x22}

	uuidDemo := UUIDDemo{
		ByteUUID:   validUUID,
		StringUUID: "550e8400-e29b-41d4-a716-446655440000",
	}

	if len(os.Args) > 1 && os.Args[1] == "spew" {
		spew.Dump(uuidDemo)
	} else {
		fmt.Println(pp.Print(uuidDemo))
	}
}

// END CAPTURE
