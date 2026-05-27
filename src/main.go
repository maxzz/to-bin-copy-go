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

	fmt.Println(ColorCyan + "==================================================" + ColorReset)
	fmt.Println(ColorCyan + ColorBold + "  DigitalPersona PM File Copier (Go Version) v" + Version + ColorReset)
	fmt.Println(ColorCyan + "==================================================" + ColorReset)
	fmt.Printf("Target Files Mode: %s (Debug or Release builds)\n", ColorGreen+mode+ColorReset)
	if configFilePath != "" {
		fmt.Printf("Config File Used:  %s\n", ColorGreen+configFilePath+ColorReset)
	}

	fmt.Println("Items to process:")
	for _, item := range items {
		if item.Name != "" {
			fmt.Printf("  - %s (Active: %t)\n", ColorGreen+item.Name+ColorReset, item.IsActive)
		} else {
			fmt.Printf("  - %s (Active: %t)\n", ColorGreen+"(unnamed)"+ColorReset, item.IsActive)
		}
	}
	fmt.Println(ColorCyan + "==================================================" + ColorReset)

	// Check for Administrator elevation on Windows
	if !isElevated() {
		fmt.Println(ColorYellow + "\nWARNING: This tool is NOT running with Administrator privileges.")
		fmt.Println("         Copying files to 'Program Files' will likely fail with 'Access is denied'.")
		fmt.Println("         Please run from an elevated command prompt (Administrator), or")
		fmt.Println("         right-click the executable and select 'Run as administrator'.")
		fmt.Println("==================================================" + ColorReset)
	}

	// Execute copy process
	ExecuteItems(items, args.IsRelease)

	fmt.Println("\nProcess complete.")
}
