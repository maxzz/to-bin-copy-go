package main

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/windows"
)

func waitForKey() {
	fd := os.Stdin.Fd()
	h := windows.Handle(fd)

	var st uint32
	if err := windows.GetConsoleMode(h, &st); err != nil {
		// Fallback to simple enter-press if not in a real console (e.g. redirected or piped)
		var buf [1]byte
		_, _ = os.Stdin.Read(buf[:])
		return
	}

	// Disable echo input, line buffering, and standard system processing of ctrl-c/backspace
	raw := st &^ (windows.ENABLE_ECHO_INPUT | windows.ENABLE_LINE_INPUT | windows.ENABLE_PROCESSED_INPUT)
	if err := windows.SetConsoleMode(h, raw); err != nil {
		var buf [1]byte
		_, _ = os.Stdin.Read(buf[:])
		return
	}
	defer windows.SetConsoleMode(h, st)

	var buf [1]byte
	_, _ = os.Stdin.Read(buf[:])
}

func ShowSuccessScreen() {
	fmt.Println()
	fmt.Println(ColorGreen + "                                                 ____    _" + ColorReset)
	fmt.Println(ColorGreen + "                                                / __ \\  | |" + ColorReset)
	fmt.Println(ColorGreen + "                                               | |  | | | | __" + ColorReset)
	fmt.Println(ColorGreen + "                                               | |  | | | |/ /" + ColorReset)
	fmt.Println(ColorGreen + "                                               | |__| | |   < " + ColorReset)
	fmt.Println(ColorGreen + "                                                \\____/  |_|\\_\\" + ColorReset)
	fmt.Println()
	fmt.Println(ColorGreen + "Everything is normal." + ColorReset)
	time.Sleep(1750 * time.Millisecond)
}

func ShowFailedScreen() {
	fmt.Println()
	fmt.Println(ColorRed + "         _..--\"\\  `|`\"\"--.._" + ColorReset)
	fmt.Println(ColorRed + "      .-'       \\  |        `'-." + ColorReset)
	fmt.Println(ColorRed + "     /           \\_|___...----'`\\" + ColorReset)
	fmt.Println(ColorRed + "    |__,,..--\"\"``(_)--..__      |" + ColorReset)
	fmt.Println(ColorRed + "    '\\     _.--'`.I._     ''--..'" + ColorReset)
	fmt.Println(ColorRed + "      `''\"`,#JGS/_|_\\###,.--'`" + ColorReset)
	fmt.Println(ColorRed + "        ,#'  _.:`___`:-._'#,              ____" + ColorReset)
	fmt.Println(ColorRed + "       #'  ,~'-;(oIo);-'~, '#            / __ \\" + ColorReset)
	fmt.Println(ColorRed + "       #   `~-(  |    )=~`  #           | |  | |  ___   _ __   ___ " + ColorReset)
	fmt.Println(ColorRed + "       #       | |_  |      #           | |  | | / _ \\ | '_ \\ / __|" + ColorReset)
	fmt.Println(ColorRed + "       #       ; ._. ;      #           | |__| || (_) || |_) |\\__ \\" + ColorReset)
	fmt.Println(ColorRed + "       #  _..-;|\\ - /|;-._  #            \\____/  \\___/ | .__/ |___/" + ColorReset)
	fmt.Println(ColorRed + "       #-'   /_ \\\\_// _\\  '-#" + ColorReset)
	fmt.Println(ColorRed + "     /`#    ; /__\\-'__\\;    #`\\" + ColorReset)
	fmt.Println(ColorRed + "    ;  #\\.--|  |O  O   |'-./#  ;" + ColorReset)
	fmt.Println(ColorRed + "    |__#/   \\ _;O__O___/   \\#__|" + ColorReset)
	fmt.Println(ColorRed + "     | #\\    [I_[_]__I]    /# |" + ColorReset)
	fmt.Println(ColorRed + "     \\_(#   ;  |O  O   ;   #)_/" + ColorReset)
	fmt.Println(ColorRed + "            |  |       |" + ColorReset)
	fmt.Println(ColorRed + "            |  ;       |" + ColorReset)
	fmt.Println(ColorRed + "            |  |       |" + ColorReset)
	fmt.Println(ColorRed + "            ;  |       |" + ColorReset)
	fmt.Println(ColorRed + "            |  |       |" + ColorReset)
	fmt.Println(ColorRed + "            |  |       ;" + ColorReset)
	fmt.Println(ColorRed + "            |  |       |" + ColorReset)
	fmt.Println(ColorRed + "            '-.;____..-'" + ColorReset)
	fmt.Println(ColorRed + "              |  ||  |" + ColorReset)
	fmt.Println(ColorRed + "              |__||__|" + ColorReset)
	fmt.Println(ColorRed + "              [__][__]" + ColorReset)
	fmt.Println(ColorRed + "            .-'-.||.-'-." + ColorReset)
	fmt.Println(ColorRed + "           (___.'  '.___)" + ColorReset)
	fmt.Println()
	fmt.Println(ColorRed + "Errors were observed during execution." + ColorReset)
	fmt.Println("Press any key to close this window...")
	waitForKey()
}
