//go:build ignore

// Nested structs and complex data structures
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/jogly/pretty"
)

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

// START CAPTURE
var pp = pretty.New().WithMaxWidth(40)

func main() {
	user := User{
		ID:      12345,
		Name:    "Ponce de Leon",
		Email:   "ponce.de.leon@clonehigh.edu",
		Active:  true,
		Created: time.Date(2020, 1, 15, 10, 30, 0, 0, time.UTC),
		Profile: UserProfile{
			Bio:       "Explorer of the New World and pants owner",
			Website:   "https://ponce.de.leon",
			Skills:    []string{"Sailing", "Pants"},
			Languages: []string{"English", "Spanish", "French"},
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
		Tags: []string{"explorer", "clone", "sailor"},
	}

	if len(os.Args) > 1 && os.Args[1] == "spew" {
		spew.Dump(user)
	} else {
		fmt.Println(pp.Print(user))
	}
}

// END CAPTURE
