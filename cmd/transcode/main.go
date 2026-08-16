package main

import (
	"fmt"
	"os"

	"example.com/golabel/transcodewrap/command"
)

func main() {
	if err := command.NewRootCommand(os.Stdout, os.Stderr).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
