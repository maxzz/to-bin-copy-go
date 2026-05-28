package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
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

var ForceCopy bool
var RestartAgent bool

// ExecuteItems executes the copy operations for the selected config items.
// Returns true if any errors occurred during execution, false if everything succeeded.
func ExecuteItems(items []ConfigItem, isRelease bool) bool {
	var bDpAgentIsDead bool
	var shouldRestart bool
	var bin32FolderToRestart string
	var hasErrors bool

	for _, item := range items {
		if item.Paths != nil {
			if item.Paths.Dp && (RestartAgent || (item.Paths.Restart != nil && *item.Paths.Restart)) {
				shouldRestart = true
			}

			sourcePaths := item.Paths.Src.Debug
			if isRelease {
				sourcePaths = item.Paths.Src.Release
			}

			dstCfg := item.Paths.Dst
			// Normalize custom destination paths
			if dstCfg.Win32 != "" {
				dstCfg.Win32 = filepath.Clean(filepath.FromSlash(dstCfg.Win32))
			}
			if dstCfg.X64 != "" {
				dstCfg.X64 = filepath.Clean(filepath.FromSlash(dstCfg.X64))
			}

			var bIsWin32, bIsWin64 bool

			for _, sourcePath := range sourcePaths {
				sourcePathClean := filepath.Clean(filepath.FromSlash(strings.TrimSpace(sourcePath)))
				sourcePathLower := strings.ToLower(sourcePathClean)

				if strings.HasSuffix(sourcePathLower, "win32") {
					bIsWin32 = true

					if !bDpAgentIsDead {
						bDpAgentIsDead = KillDpAgent()
					}

					bin32Folder := dstCfg.Win32
					if bin32Folder == "" {
						programs32Folder := os.Getenv("ProgramFiles(x86)")
						if programs32Folder == "" {
							programs32Folder = `C:\Program Files (x86)`
						}
						bin32Folder = filepath.Join(programs32Folder, `DigitalPersona\Bin`)
					}
					bin32FolderToRestart = bin32Folder

					fmt.Printf("Processing %s -> %s\n", sourcePathClean, bin32Folder)

					if item.Paths.Dp {
						for _, sFileName := range win32Files {
							if !CopyFileToBin(sourcePathClean, sFileName, bin32Folder) {
								hasErrors = true
							}
						}
					} else {
						// Custom include/exclude regex matching
						customFailed, err := copyCustomFiles(sourcePathClean, bin32Folder, item.Paths.SrcFilesInclude, item.Paths.SrcFilesExclude)
						if err != nil {
							fmt.Fprintf(os.Stderr, ColorRed+"  Error copying custom files: %v\n"+ColorReset, err)
							hasErrors = true
						} else if customFailed {
							hasErrors = true
						}
					}

				} else if strings.HasSuffix(sourcePathLower, "x64") {
					bIsWin64 = true

					if !bDpAgentIsDead {
						bDpAgentIsDead = KillDpAgent()
					}

					bin64Folder := dstCfg.X64
					if bin64Folder == "" {
						programs64Folder := os.Getenv("ProgramFiles")
						if programs64Folder == "" {
							programs64Folder = `C:\Program Files`
						}
						bin64Folder = filepath.Join(programs64Folder, `DigitalPersona\Bin`)
					}

					fmt.Printf("Processing %s -> %s\n", sourcePathClean, bin64Folder)

					if item.Paths.Dp {
						for _, sFileName := range x64Files {
							if !CopyFileToBin(sourcePathClean, sFileName, bin64Folder) {
								hasErrors = true
							}
						}
					} else {
						// Custom include/exclude regex matching
						customFailed, err := copyCustomFiles(sourcePathClean, bin64Folder, item.Paths.SrcFilesInclude, item.Paths.SrcFilesExclude)
						if err != nil {
							fmt.Fprintf(os.Stderr, ColorRed+"  Error copying custom files: %v\n"+ColorReset, err)
							hasErrors = true
						} else if customFailed {
							hasErrors = true
						}
					}
				}
			}

			if !bIsWin32 && !bIsWin64 && len(sourcePaths) > 0 {
				fmt.Println(ColorYellow + "Warning: No valid source paths found ending with 'Win32' or 'x64'." + ColorReset)
			}
		}

		if len(item.Files) > 0 {
			fmt.Printf("Processing individual files list...\n")
			for _, fileBlock := range item.Files {
				if fileBlock.Src == "" || fileBlock.Dst == "" {
					continue
				}

				if !bDpAgentIsDead {
					// Preemptively kill DPAgent if any file is copied
					bDpAgentIsDead = KillDpAgent()
				}

				srcClean := filepath.Clean(filepath.FromSlash(strings.TrimSpace(fileBlock.Src)))
				dstClean := filepath.Clean(filepath.FromSlash(strings.TrimSpace(fileBlock.Dst)))

				if !CopyFileWithChecks(srcClean, dstClean, ForceCopy) {
					hasErrors = true
				}
			}
		}
	}

	if shouldRestart {
		if bin32FolderToRestart == "" {
			programs32Folder := os.Getenv("ProgramFiles(x86)")
			if programs32Folder == "" {
				programs32Folder = `C:\Program Files (x86)`
			}
			bin32FolderToRestart = filepath.Join(programs32Folder, `DigitalPersona\Bin`)
		}
		if err := RestartDpAgent(bin32FolderToRestart); err != nil {
			fmt.Fprintf(os.Stderr, ColorYellow+"Warning: Could not restart DPAgent: %v\n"+ColorReset, err)
			hasErrors = true
		}
	}

	return hasErrors
}

// CopyPmFilesToBin is a legacy wrapper for backwards compatibility with tests and older invocation patterns.
func CopyPmFilesToBin(sourcePaths []string, dstCfg DstConfig) {
	item := ConfigItem{
		SetName:  "Legacy Run",
		IsActive: true,
		Paths: &PathBlock{
			Dp: true,
			Src: SrcConfig{
				Debug:   sourcePaths,
				Release: sourcePaths,
			},
			Dst: dstCfg,
		},
	}
	ExecuteItems([]ConfigItem{item}, false)
}

func copyCustomFiles(sourceDir, destDir string, includes, excludes []string) (bool, error) {
	// Compile regex patterns first
	var includeRegexes []*regexp.Regexp
	for _, p := range includes {
		re, err := regexp.Compile(p)
		if err != nil {
			return false, fmt.Errorf("invalid include pattern %q: %v", p, err)
		}
		includeRegexes = append(includeRegexes, re)
	}

	var excludeRegexes []*regexp.Regexp
	for _, p := range excludes {
		re, err := regexp.Compile(p)
		if err != nil {
			return false, fmt.Errorf("invalid exclude pattern %q: %v", p, err)
		}
		excludeRegexes = append(excludeRegexes, re)
	}

	anyFailed := false
	// Walk directory recursively
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Get relative path from sourceDir to this file
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Normalize backslashes to forward slashes for matching patterns
		relPathSlash := filepath.ToSlash(relPath)

		// Match includes (if empty, matches all)
		included := len(includeRegexes) == 0
		for _, re := range includeRegexes {
			if re.MatchString(relPathSlash) {
				included = true
				break
			}
		}

		// Match excludes
		excluded := false
		for _, re := range excludeRegexes {
			if re.MatchString(relPathSlash) {
				excluded = true
				break
			}
		}

		if included && !excluded {
			targetPath := filepath.Join(destDir, relPath)
			if !CopyFileWithChecks(path, targetPath, ForceCopy) {
				anyFailed = true
			}
		}

		return nil
	})

	return anyFailed, err
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
// Returns true on success, false on error.
func CopyFileToBin(sourcePath, sFileName, sDestPath string) bool {
	return CopyFileWithChecks(filepath.Join(sourcePath, sFileName), filepath.Join(sDestPath, sFileName), ForceCopy)
}

// CopyFileWithChecks copies a single file with force/timestamp checks and in-use renaming fallback.
// Returns true on success, false on error.
func CopyFileWithChecks(srcPath, dstPath string, force bool) bool {
	sourceFileInfo, err := os.Stat(srcPath)
	if os.IsNotExist(err) {
		fmt.Println(ColorRed + "  No source file!!!: " + srcPath + ColorReset)
		return false
	} else if err != nil {
		fmt.Fprintln(os.Stderr, ColorRed+fmt.Sprintf("  Error reading source file %s: %v", srcPath, err)+ColorReset)
		return false
	}

	sourceFileTime := sourceFileInfo.ModTime().UTC()
	bTimesOK := true

	if !force {
		destFileInfo, err := os.Stat(dstPath)
		if err == nil {
			destFileTime := destFileInfo.ModTime().UTC()
			bTimesOK = sourceFileTime.After(destFileTime)
		}
	}

	if bTimesOK {
		sDestDir := filepath.Dir(dstPath)
		sFileName := filepath.Base(dstPath)
		for {
			err := doCopy(srcPath, dstPath)
			if err == nil {
				// Success, set the destination file time to match the source file time
				_ = os.Chtimes(dstPath, sourceFileInfo.ModTime(), sourceFileInfo.ModTime())
				fmt.Println(ColorGreen + "  Copied " + dstPath + ColorReset)
				return true
			}

			if isSharingViolation(err) {
				fmt.Println(ColorYellow + "  File in use: " + dstPath + ColorReset)
				renameErr := RenameDestFile(sDestDir, sFileName)
				if renameErr != nil {
					fmt.Fprintln(os.Stderr, ColorRed+fmt.Sprintf("  Failed to rename locked file: %v", renameErr)+ColorReset)
					return false
				}
			} else {
				fmt.Fprintln(os.Stderr, ColorRed+fmt.Sprintf("  Failed copying file %s, error: %v", dstPath, err)+ColorReset)
				return false
			}
		}
	} else {
		fmt.Println(ColorDim + "  Same time, skipping: " + filepath.Base(dstPath) + ColorReset)
		return true
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
