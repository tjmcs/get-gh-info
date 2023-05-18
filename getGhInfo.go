/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/tjmcs/get-gh-info/cmd"
	_ "github.com/tjmcs/get-gh-info/cmd/repo"
	_ "github.com/tjmcs/get-gh-info/cmd/repo/issues"
	_ "github.com/tjmcs/get-gh-info/cmd/repo/pulls"
	_ "github.com/tjmcs/get-gh-info/cmd/user"
)

// start the program by running the command passed in via the CLI
func main() {
	cmd.Execute()
}
