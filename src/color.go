package main

import (
	"os"

	"golang.org/x/sys/windows"
)

const (
	ColorReset  = "\x1b[0m"
	ColorBold   = "\x1b[1m"
	ColorDim    = "\x1b[90m" // Decreased brightness / Gray
	ColorRed    = "\x1b[31m"
	ColorGreen  = "\x1b[32m"
	ColorYellow = "\x1b[33m"
	ColorCyan   = "\x1b[36m"
)

// initConsole enables Virtual Terminal Processing in Windows Console to support ANSI colors.
func initConsole() {
	stdout := windows.Handle(os.Stdout.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(stdout, &mode); err == nil {
		mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
		_ = windows.SetConsoleMode(stdout, mode)
	}

	stderr := windows.Handle(os.Stderr.Fd())
	if err := windows.GetConsoleMode(stderr, &mode); err == nil {
		mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
		_ = windows.SetConsoleMode(stderr, mode)
	}
}
