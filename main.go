package main

import "github.com/pmaojo/goploy/cmd"

// main is the entry point of the application.
// It delegates execution to the root command in the cmd package.
func main() {
	cmd.Execute()
}
