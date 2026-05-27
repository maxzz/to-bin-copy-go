# DigitalPersona PM File Copier (Go Version)

A robust, lightweight command-line utility written in Go for copying DigitalPersona PM native binaries and driver files from build output directories to the system's DigitalPersona installation directory.

This utility is a direct conversion of the original C# WinForms application `CopyPmFilesToBin` to Go. It keeps all of the original functionality intact, adding a cleaner command-line interface, detailed progress logs, and helpful system privilege warnings.

---

## Key Features

1. **Flexible Configuration Resolution (Hierarchical Fallback)**:
   - **CLI Source flag (`-source`)**: Manually pass a comma-separated list of paths to bypass configuration files completely.
   - **Custom Config File (`-config <path>`)**: Point the tool to a custom JSON configuration file.
   - **Default Config File (`config.json`)**: Looks for a `config.json` file in the current working directory, and falls back to looking next to the executable.
   - **Windows Registry Compatibility**: If no config files are found or configured paths are empty, it automatically reads the registry key used by the C# application (`HKEY_CURRENT_USER\SOFTWARE\SergeiMenchenin\CopyPmFilesToBin`), making it fully backward-compatible with your existing registry setup!

2. **Graceful DPAgent Termination**:
   - Uses native Windows Win32 API (`FindWindowW`, `PostMessageW`, and `WaitForSingleObject`) via `golang.org/x/sys/windows` to gracefully request `DPAgent.exe` to close, and waits for its exit before performing the file copy.

3. **Intelligent Timestamp Checking**:
   - Compares Last Modified UTC times between the source and destination files. Files are only copied if the source file is strictly newer, saving writes and avoiding unnecessary copying.

4. **Locked File Handling (File-In-Use Renaming)**:
   - If copying fails because a file is locked or in use (Sharing Violation error `0x80070020`), the tool automatically renames the existing locked destination file (e.g., `DpFbView.dll` -> `DpFbView_1.dll`, `DpFbView_2.dll`, etc.) and retries the copy successfully.

5. **Administrator Privilege Warning**:
   - Since copying to `C:\Program Files` or `C:\Program Files (x86)` requires elevated privileges on Windows, the tool checks for administrator rights on startup and outputs a user-friendly warning if it's run as a standard user.

---

## Folder Structure

- `main.go`: Entry point and program orchestration.
- `config.go`: Argument parsing, JSON configuration, and Windows Registry integration.
- `process.go`: Graceful Win32 process search and close logic for `DPAgent.exe`.
- `copy.go`: Core copy loop, timestamp comparison, and locked file renaming.
- `config.json`: Sample configuration file template.

---

## File Copies Performed

Based on the suffix of each source path, the tool copies the appropriate architecture-specific files:

### Win32 Source Paths (ending with `Win32`)
Copies to `C:\Program Files (x86)\DigitalPersona\Bin`:
- `DpAgent.exe`
- `DpFbView.dll`
- `DpOFeedb.dll`
- `DpoPS.dll`
- `DpoSet.dll`
- `DPPMAdminConsole.exe`
- `DpoSetA.dll`
- `DpoTrain.dll`
- `DpoTrainMgr.dll`
- `DpStgCat.dll`

### x64 Source Paths (ending with `x64`)
Copies to `C:\Program Files\DigitalPersona\Bin`:
- `DpAgentOtsPlugin.dll`
- `DpAgentOtsPlugin.WebSdk.dll`
- `DpFbView.dll`
- `DpImporter.dll`
- `DpMiniDS.dll`
- `DpOCache.dll`
- `DpOFeedb.dll`
- `DpOnlineIDs.dll`
- `DpoPS.dll`
- `DpoSet.dll`
- `DpOtsMsg.dll`
- `DpUtt.dll`
- `DsDashboard.dll`

---

## How to Build

First, make sure you have Go installed on your machine.

1. Install dependencies (specifically the Windows system bindings package):
   ```bash
   go get golang.org/x/sys/windows
   ```

2. Compile the binary:
   ```bash
   go build -o copy-pm-files.exe
   ```

---

## How to Run

### Command-Line Arguments

- `-release`: Run in Release mode (looks for Release paths instead of Debug paths). By default, it runs in Debug mode.
- `-config <path>`: Path to a JSON configuration file (defaults to `config.json`).
- `-source <paths>`: Comma-separated list of paths (e.g. `-source "C:\src\Win32,C:\src\x64"`), bypassing any registry or config file settings.

### Examples

**1. Standard Run (Debug, uses registry fallback or local `config.json`):**
```bash
# Run with administrator privileges
.\copy-pm-files.exe
```

**2. Run in Release Mode:**
```bash
.\copy-pm-files.exe -release
```

**3. Run with Custom Paths directly (bypassing configs):**
```bash
.\copy-pm-files.exe -source "C:\MySources\Debug.Win32,C:\MySources\Debug.x64"
```

**4. Run with a Custom Configuration File:**
```bash
.\copy-pm-files.exe -config C:\Users\Public\my_custom_config.json
```

---

## Configuration File Format

If you prefer using a `config.json` file rather than command-line arguments or the Windows Registry, you can place a `config.json` next to your executable. Here is the format:

```json
{
  "sourcePathsDebug": [
    "C:\\y\\c\\dp\\pm-native\\src\\~Output\\Debug.Win32",
    "C:\\y\\c\\dp\\pm-native\\src\\~Output\\Debug.x64"
  ],
  "sourcePathsRelease": [
    "C:\\y\\c\\dp\\pm-native\\src\\~Output\\Release.Win32",
    "C:\\y\\c\\dp\\pm-native\\src\\~Output\\Release.x64"
  ]
}
```
*(Note: Backslashes in paths must be escaped as `\\` inside JSON).*
