package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

func isElevated() bool {
	token := windows.GetCurrentProcessToken()
	return token.IsElevated()
}

func main() {
	// Initialize Windows Console for Virtual Terminal (ANSI color) processing
	initConsole()

	// Parse CLI arguments
	args := parseArgs()

	// Initialize the global force copy flag
	ForceCopy = args.Force

	// Resolve the config items to process
	items, configFilePath, err := ResolveConfigAndItems(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, ColorRed+"Error: %v\n\n"+ColorReset, err)
		PrintHelp()
		os.Exit(1)
	}

	mode := "Debug"
	if args.IsRelease {
		mode = "Release"
	}

	// Print startup information, configuration, items, and elevation warnings
	printAccepted(items, mode, configFilePath)

	// Execute copy process
	ExecuteItems(items, args.IsRelease)

	fmt.Println("\nProcess complete.")
}
