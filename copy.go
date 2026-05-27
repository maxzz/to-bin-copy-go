package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

var win32Files = []string{
	"DpAgent.exe",
	"DpFbView.dll",
	"DpOFeedb.dll",
	"DpoPS.dll",
	"DpoSet.dll",
	"DPPMAdminConsole.exe",
	"DpoSetA.dll",
	"DpoTrain.dll",
	"DpoTrainMgr.dll",
	"DpStgCat.dll",
}

var x64Files = []string{
	"DpAgentOtsPlugin.dll",
	"DpAgentOtsPlugin.WebSdk.dll",
	"DpFbView.dll",
	"DpImporter.dll",
	"DpMiniDS.dll",
	"DpOCache.dll",
	"DpOFeedb.dll",
	"DpOnlineIDs.dll",
	"DpoPS.dll",
	"DpoSet.dll",
	"DpOtsMsg.dll",
	"DpUtt.dll",
	"DsDashboard.dll",
}

// CopyPmFilesToBin executes the copy operations from source paths to the target system folders.
func CopyPmFilesToBin(sourcePaths []string) {
	var bIsWin32, bIsWin64 bool
	var bDpAgentIsDead bool

	for _, sourcePath := range sourcePaths {
		sourcePathLower := strings.ToLower(sourcePath)

		if strings.HasSuffix(sourcePathLower, "win32") {
			bIsWin32 = true

			if !bDpAgentIsDead {
				bDpAgentIsDead = KillDpAgent()
			}

			programs32Folder := os.Getenv("ProgramFiles(x86)")
			if programs32Folder == "" {
				programs32Folder = `C:\Program Files (x86)`
			}
			bin32Folder := filepath.Join(programs32Folder, `DigitalPersona\Bin`)

			fmt.Printf("From %s to %s\n", sourcePath, bin32Folder)

			for _, sFileName := range win32Files {
				CopyFileToBin(sourcePath, sFileName, bin32Folder)
			}

		} else if strings.HasSuffix(sourcePathLower, "x64") {
			bIsWin64 = true

			if !bDpAgentIsDead {
				bDpAgentIsDead = KillDpAgent()
			}

			programs64Folder := os.Getenv("ProgramFiles")
			if programs64Folder == "" {
				programs64Folder = `C:\Program Files`
			}
			bin64Folder := filepath.Join(programs64Folder, `DigitalPersona\Bin`)

			fmt.Printf("From %s to %s\n", sourcePath, bin64Folder)

			for _, sFileName := range x64Files {
				CopyFileToBin(sourcePath, sFileName, bin64Folder)
			}
		}
	}

	if !bIsWin32 && !bIsWin64 {
		fmt.Println("Warning: No valid source paths found ending with 'Win32' or 'x64'.")
	}
}

// isSharingViolation checks if the error is a Windows sharing violation (ERROR_SHARING_VIOLATION = 32).
func isSharingViolation(err error) bool {
	if err == nil {
		return false
	}
	if errno, ok := err.(syscall.Errno); ok && errno == windows.ERROR_SHARING_VIOLATION {
		return true
	}
	// Try unwrapping if it is an os.PathError
	if pathErr, ok := err.(*os.PathError); ok {
		if errno, ok := pathErr.Err.(syscall.Errno); ok && errno == windows.ERROR_SHARING_VIOLATION {
			return true
		}
	}
	return false
}

// CopyFileToBin copies a single file if it's newer, and renames on sharing violation.
func CopyFileToBin(sourcePath, sFileName, sDestPath string) {
	sFullSourcePath := filepath.Join(sourcePath, sFileName)
	sFullDestPath := filepath.Join(sDestPath, sFileName)

	sourceFileInfo, err := os.Stat(sFullSourcePath)
	if os.IsNotExist(err) {
		fmt.Printf("  No source file!!!: %s\n", sFullDestPath)
		return
	} else if err != nil {
		fmt.Printf("  Error reading source file %s: %v\n", sFullSourcePath, err)
		return
	}

	sourceFileTime := sourceFileInfo.ModTime().UTC()
	bTimesOK := true

	destFileInfo, err := os.Stat(sFullDestPath)
	if err == nil {
		destFileTime := destFileInfo.ModTime().UTC()
		bTimesOK = sourceFileTime.After(destFileTime)
	}

	if bTimesOK {
		for {
			err := doCopy(sFullSourcePath, sFullDestPath)
			if err == nil {
				// Success, set the destination file time to match the source file time
				_ = os.Chtimes(sFullDestPath, sourceFileInfo.ModTime(), sourceFileInfo.ModTime())
				fmt.Printf("  Copied %s\n", sFullDestPath)
				break
			}

			if isSharingViolation(err) {
				fmt.Printf("  File in use: %s\n", sFullDestPath)
				renameErr := RenameDestFile(sDestPath, sFileName)
				if renameErr != nil {
					fmt.Printf("  Failed to rename locked file: %v\n", renameErr)
					break
				}
			} else {
				fmt.Printf("  Failed copying file %s, error: %v\n", sFullDestPath, err)
				break
			}
		}
	} else {
		fmt.Printf("  Same time, skipping: %s\n", sFileName)
	}
}

// doCopy copies file contents from src to dst.
func doCopy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// Ensure destination directory exists before copying
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Sync()
}

// RenameDestFile renames a locked destination file by appending _1, _2, etc.
func RenameDestFile(sDestPath, sFileName string) error {
	ext := filepath.Ext(sFileName)
	fileNameOnly := strings.TrimSuffix(sFileName, ext)

	sSearchPattern := filepath.Join(sDestPath, fileNameOnly+"_*"+ext)
	renamedFiles, err := filepath.Glob(sSearchPattern)
	if err != nil {
		return err
	}

	fullOldFileName := filepath.Join(sDestPath, sFileName)
	var newFileName string

	if len(renamedFiles) == 0 {
		newFileName = fileNameOnly + "_1" + ext
	} else {
		maxNumber := 0
		for _, fPath := range renamedFiles {
			fName := filepath.Base(fPath)
			fileNameWithNumber := strings.TrimSuffix(fName, ext)
			numOnly := strings.Replace(fileNameWithNumber, fileNameOnly+"_", "", 1)

			if num, err := strconv.Atoi(numOnly); err == nil {
				if num > maxNumber {
					maxNumber = num
				}
			}
		}
		maxNumber++
		newFileName = fmt.Sprintf("%s_%d%s", fileNameOnly, maxNumber, ext)
	}

	fullNewFileName := filepath.Join(sDestPath, newFileName)
	return os.Rename(fullOldFileName, fullNewFileName)
}
