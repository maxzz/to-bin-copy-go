package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32           = windows.NewLazySystemDLL("user32.dll")
	procFindWindowW  = user32.NewProc("FindWindowW")
	procPostMessageW = user32.NewProc("PostMessageW")
)

const (
	WM_CLOSE = 0x0010
)

func findWindow(className, windowName string) (windows.HWND, error) {
	var classPtr *uint16
	var windowPtr *uint16
	var err error

	if className != "" {
		classPtr, err = windows.UTF16PtrFromString(className)
		if err != nil {
			return 0, err
		}
	}
	if windowName != "" {
		windowPtr, err = windows.UTF16PtrFromString(windowName)
		if err != nil {
			return 0, err
		}
	}

	ret, _, _ := procFindWindowW.Call(
		uintptr(unsafe.Pointer(classPtr)),
		uintptr(unsafe.Pointer(windowPtr)),
	)
	return windows.HWND(ret), nil
}

func postMessage(hWnd windows.HWND, msg uint32, wParam, lParam uintptr) bool {
	ret, _, _ := procPostMessageW.Call(
		uintptr(hWnd),
		uintptr(msg),
		wParam,
		lParam,
	)
	return ret != 0
}

func findProcessIDsByName(name string) ([]uint32, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)

	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))

	err = windows.Process32First(snapshot, &pe)
	if err != nil {
		return nil, err
	}

	var pids []uint32
	target := strings.ToLower(name)
	for {
		exeName := strings.ToLower(windows.UTF16ToString(pe.ExeFile[:]))
		if exeName == target || exeName == target+".exe" {
			pids = append(pids, pe.ProcessID)
		}

		err = windows.Process32Next(snapshot, &pe)
		if err != nil {
			break
		}
	}

	return pids, nil
}

// KillDpAgent closes the DigitalPersona Agent process gracefully.
// Returns true if the agent is dead or was successfully terminated,
// false if terminating failed.
func KillDpAgent() bool {
	pids, err := findProcessIDsByName("DPAgent")
	if err != nil {
		fmt.Printf("Error listing processes: %v\n", err)
		return false
	}

	if len(pids) == 0 {
		return true // No processes running, we're good
	}

	for _, pid := range pids {
		hProcess, err := windows.OpenProcess(windows.SYNCHRONIZE, false, pid)
		if err != nil {
			continue
		}
		defer windows.CloseHandle(hProcess)

		const lpClassName = "DigitalPersona Pro5.x Agent Window Class"
		hWnd, err := findWindow(lpClassName, "")
		if err == nil && hWnd != 0 {
			fmt.Printf("Found DPAgent (PID %d) window, sending WM_CLOSE...\n", pid)
			postMessage(hWnd, WM_CLOSE, 0, 0)

			// Wait up to 10 seconds
			event, err := windows.WaitForSingleObject(hProcess, 10000)
			if err == nil && event == windows.WAIT_OBJECT_0 {
				fmt.Println("DPAgent terminated gracefully.")
				time.Sleep(1 * time.Second)
				return true
			}
		}
	}

	// Sleep 3 seconds if graceful close failed or window was not found
	fmt.Println("Could not gracefully terminate DPAgent via window close.")
	time.Sleep(3 * time.Second)
	return false
}

// RestartDpAgent starts the DigitalPersona Agent process from the specified binary directory.
func RestartDpAgent(binDir string) error {
	agentPath := filepath.Join(binDir, "DpAgent.exe")
	fmt.Printf("Restarting DPAgent from: %s...\n", agentPath)
	cmd := exec.Command(agentPath)
	cmd.Dir = binDir
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start DPAgent.exe: %v", err)
	}
	fmt.Println(ColorGreen + "DPAgent.exe restarted successfully." + ColorReset)
	return nil
}
