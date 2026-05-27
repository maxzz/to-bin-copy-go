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
	// Parse CLI arguments
	args := parseArgs()

	// Get resolved source paths and destination configurations
	paths, dstCfg, mode, sourceUsed, err := GetSourcePaths(args)
	if err != nil {
		fmt.Printf("Error: %v\n\n", err)
		fmt.Println("Usage of copy-pm-files-to-bin:")
		fmt.Println("  to-bin-copy-go [flags]")
		fmt.Println("\nFlags:")
		fmt.Println("  -release")
		fmt.Println("        Run in Release mode (default is Debug mode)")
		fmt.Println("  -config string")
		fmt.Println("        Path to the configuration JSON file (default \"config.json\")")
		fmt.Println("  -source string")
		fmt.Println("        Comma-separated list of custom source directories (bypasses config file/registry)")
		os.Exit(1)
	}

	fmt.Println("==================================================")
	fmt.Println("  DigitalPersona PM File Copier (Go Version)")
	fmt.Println("==================================================")
	fmt.Printf("Target Files Mode: %s (Debug or Release builds)\n", mode)
	fmt.Printf("Source Resolved:   %s\n", sourceUsed)
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
	fmt.Println("==================================================")

	// Check for Administrator elevation on Windows
	if !isElevated() {
		fmt.Println("\nWARNING: This tool is NOT running with Administrator privileges.")
		fmt.Println("         Copying files to 'Program Files' will likely fail with 'Access is denied'.")
		fmt.Println("         Please run from an elevated command prompt (Administrator), or")
		fmt.Println("         right-click the executable and select 'Run as administrator'.")
		fmt.Println("==================================================")
	}

	// Execute copy process
	CopyPmFilesToBin(paths, dstCfg)

	fmt.Println("\nProcess complete.")
}
