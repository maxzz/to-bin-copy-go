package main

import (
	"fmt"
	"os"
)

// PrintError prints a formatted error message in Yellow to stderr,
// followed by the usage help suggestion.
func PrintError(msg string) {
	fmt.Fprintln(os.Stderr, ColorYellow+"Error: "+msg+ColorReset)
	fmt.Println()
	fmt.Fprintln(os.Stderr, ColorDim+"Run 'copy-pm-files.exe -help' to see full usage instructions."+ColorReset)
}

// PrintWarning prints a formatted warning message in Yellow to stderr,
// followed by the usage help suggestion.
func PrintWarning(msg string) {
	fmt.Fprintln(os.Stderr, ColorYellow+"Warning: "+msg+ColorReset)
	fmt.Println()
	fmt.Fprintln(os.Stderr, ColorDim+"Run 'copy-pm-files.exe -help' to see full usage instructions."+ColorReset)
}
