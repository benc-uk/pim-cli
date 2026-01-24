package main

import "github.com/benc-uk/pimg-cli/cmd"

var version = "0.0.0"

func main() {
	_ = cmd.Execute(version)
}
