package main

import (
	"fmt"
	"os"

	"github.com/inetmanageai/mai/backend"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		panic("please send args:\nmai <command> <arguments>")
	}

	// select command
	switch args[1] {
	case "create":
		if len(args) < 3 {
			panic("please send args:\nmai create <arguments>")
		}

		// select create command
		switch args[2] {
		case "backend":
			if args[2] == "backend" {
				ctx := backend.NewBackend(nil)
				err := ctx.Create()
				if err != nil {
					panic(err)
				}
			}

		default:
			panic(fmt.Sprintf("command %s not found", args[1]))
		}

	default:
		panic(fmt.Sprintf("command %s not found", args[1]))
	}
}
