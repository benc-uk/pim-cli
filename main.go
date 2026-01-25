// ==============================================================================================
// CLI entry point, standard for a Cobra-based application
// See cmd/root.go for main command implementation
// ===============================================================================================

package main

import "github.com/benc-uk/pimg-cli/cmd"

var version = "0.0.0"

func main() {
	_ = cmd.Execute(version)
}
