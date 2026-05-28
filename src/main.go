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
	RestartAgent = args.Restart

	// Resolve the config items to process
	items, configFilePath, err := ResolveConfigAndItems(args)
	if err != nil {
		PrintError(err.Error())
		os.Exit(1)
	}

	mode := "Debug"
	if args.IsRelease {
		mode = "Release"
	}

	// Print startup information, configuration, items, and elevation warnings
	printAccepted(items, mode, configFilePath)

	// Execute copy process
	hasErrors := ExecuteItems(items, args.IsRelease)

	fmt.Println("\nProcess complete.")

	// Determine if wait is enabled
	waitEnabled := false
	if args.WaitSpecified {
		waitEnabled = args.Wait
	} else {
		for _, item := range items {
			if item.Wait != nil && *item.Wait {
				waitEnabled = true
				break
			}
		}
	}

	if waitEnabled {
		if hasErrors {
			ShowFailedScreen()
		} else {
			ShowSuccessScreen()
		}
	}
}
