package main

import (
	"fmt"
	"os"

	"github.com/algolia/harvestcli/cmd"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: harvestcli <command> [<args>]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "convert-csv":
		cmd := cmd.NewConvertToCSVCommand(os.Args[2:])
		cmd.Run()
		cmd.Close()
	case "convert-json":
		cmd := cmd.NewConvertToJSONCommand(os.Args[2:])
		cmd.Run()
		cmd.Close()
	case "associate":
		cmd := cmd.NewAssociateCommand(os.Args[2:])
		cmd.Run()
		cmd.Close()
	case "merge":
		cmd := cmd.NewMergeCommand(os.Args[2:])
		cmd.Run()
		cmd.Close()
	default:
		fmt.Printf("%q: not a valid command\n", os.Args[1])
		os.Exit(1)
	}
}
