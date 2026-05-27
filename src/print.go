package main

import "fmt"

// printAccepted prints the startup banners, configured items, and checks for administrator elevation.
func printAccepted(items []ConfigItem, mode, configFilePath string) {
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
}
