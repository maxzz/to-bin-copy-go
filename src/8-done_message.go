package main

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows"
)

var (
	msvcrt    = windows.NewLazySystemDLL("msvcrt.dll")
	procGetch = msvcrt.NewProc("_getch")
)

func waitForKey() {
	_, _, _ = procGetch.Call()
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
