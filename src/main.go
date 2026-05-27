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

	// Get resolved source paths and destination configurations
	paths, dstCfg, mode, sourceUsed, err := GetSourcePaths(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, ColorRed+"Error: %v\n\n"+ColorReset, err)
		PrintHelp()
		os.Exit(1)
	}

	fmt.Println(ColorCyan + "==================================================" + ColorReset)
	fmt.Println(ColorCyan + ColorBold + "  DigitalPersona PM File Copier (Go Version)" + ColorReset)
	fmt.Println(ColorCyan + "==================================================" + ColorReset)
	fmt.Printf("Target Files Mode: %s (Debug or Release builds)\n", ColorGreen+mode+ColorReset)
	fmt.Printf("Source Resolved:   %s\n", ColorGreen+sourceUsed+ColorReset)
	fmt.Println("Source Paths:")
	for _, p := range paths {
		fmt.Printf("  - %s\n", p)
	}
	if dstCfg.Win32 != "" || dstCfg.X64 != "" {
		fmt.Println("Custom Destination Paths:")
		if dstCfg.Win32 != "" {
			fmt.Printf("  - Win32: %s\n", dstCfg.Win32)
		}
		if dstCfg.X64 != "" {
			fmt.Printf("  - x64:   %s\n", dstCfg.X64)
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
	CopyPmFilesToBin(paths, dstCfg)

	fmt.Println("\nProcess complete.")
}
