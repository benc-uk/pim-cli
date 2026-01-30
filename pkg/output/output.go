// =====================================================================
// Pretty basic console/stdout output package with two verbosity levels
// =====================================================================

package output

import (
	"fmt"
	"os"
)

// Level represents the verbosity level for output
type Level int

const (
	// Quiet - only errors and essential output
	Quiet Level = iota
	// Normal - standard output (default)
	Normal
)

var currentLevel = Normal

// SetLevel sets the global output verbosity level
func SetLevel(level Level) {
	currentLevel = level
}

// GetLevel returns the current output verbosity level
func GetLevel() Level {
	return currentLevel
}

// Println outputs a message at Normal level with newline
func Println(args ...any) {
	if currentLevel >= Normal {
		fmt.Println(args...)
	}
}

// Printf outputs a formatted message at Normal level
func Printf(format string, args ...any) {
	if currentLevel >= Normal {
		fmt.Printf(format, args...)
	}
}

// Error outputs an error message (always shown, even in Quiet mode)
func Error(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}

// Printlnq outputs a message that is always shown (even in Quiet mode)
func Printlnq(args ...any) {
	fmt.Println(args...)
}

// Printfq outputs a formatted message that is always shown (even in Quiet mode)
func Printfq(format string, args ...any) {
	fmt.Printf(format, args...)
}

// Fatalf outputs an error message and exits with code 1
func Fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "\033[31mError: "+format+"\033[0m", args...)
	os.Exit(1)
}
