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
   - Since copying to `C:/Program Files` or `C:/Program Files (x86)` requires elevated privileges on Windows, the tool checks for administrator rights on startup and outputs a user-friendly warning if it's run as a standard user.

---

## Folder Structure

- `src/`: Directory containing all Go source files.
  - `main.go`: Entry point and program orchestration.
  - `config.go`: Argument parsing, JSON configuration, and Windows Registry integration.
  - `help.go`: Modular help output displaying usage, default options, and behaviors.
  - `process.go`: Graceful Win32 process search and close logic for `DPAgent.exe`.
  - `copy.go`: Core copy loop, timestamp comparison, and locked file renaming.
  - `copy_test.go`: Unit tests for file utility behaviors.
- `tests/`: Directory containing various preconfigured examples of JSON configuration files.
- `config.json`: Sample configuration file template.
- `package.json`: NPM package manifest for unified scripts (run, test, build).
- `.gitignore`: Configured to ignore Go/Node/Windows artifacts and generated binaries.

---

## How to Build and Run (Using NPM/Node scripts)

You can run the project using standard `npm` commands or the direct `go` commands below.

### Using NPM Scripts

- **Run in Debug Mode (Default):**
  ```bash
  npm start
  ```
- **Run in Release Mode:**
  ```bash
  npm run start:release
  ```
- **Build the executable:**
  ```bash
  npm run build
  ```
  This creates the compiled binary in `bin/copy-pm-files.exe`.
- **Run unit tests:**
  ```bash
  npm test
  ```

---

## How to Build (Using Native Go Commands)

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

- `-release`: Run the utility targeting **Release** compiled files (looks for paths specified in the `sourcePathsRelease` array instead of `sourcePathsDebug`). By default, the utility operates on **Debug** compiled files.
- `-config <path>`: Path to a JSON configuration file (defaults to `config.json`).
- `-source <paths>`: Comma-separated list of paths (e.g. `-source "C:/src/Win32,C:/src/x64"`), bypassing any registry or config file settings.

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
.\copy-pm-files.exe -source "C:/MySources/Debug.Win32,C:/MySources/Debug.x64"
```

**4. Run with a Custom Configuration File:**
```bash
.\copy-pm-files.exe -config C:/Users/Public/my_custom_config.json
```

---

## Detailed Configuration & Target Mode Clarification

### Important: Target Modes vs. Utility Compilation
It is crucial to clarify that the **Debug** and **Release** target modes **do not describe how this utility itself is compiled**, nor does running with `-release` change the performance characteristics of this Go tool. Rather:
* **The mode dictates which sets of your project's compiled binary files are being targeted for copying.**
* When running in **Debug mode** (default), the utility reads the paths where your C++/C# compiler outputs the **Debug target builds** (typically ending with `Debug.Win32` and `Debug.x64`).
* When running in **Release mode** (using `-release`), the utility targets the paths where your compiler outputs the **Release target builds** (typically ending with `Release.Win32` and `Release.x64`).

---

## Configuration File Format & Structure

If you prefer using a `config.json` file rather than command-line arguments or the Windows Registry, you can place a `config.json` next to your executable.

### Flexible Path Syntax (Slashes & Normalization)
- **Mixed Slashes**: Both forward slashes (`/`) and backslashes (`\`) are fully supported.
- **Forward Slashes**: Use `/` in configuration paths (recommended; no escaping required in JSON).
- **Backslashes**: Windows-style `\` paths are also accepted at runtime and normalized automatically.
- **Normalization**: Paths are automatically cleaned to match the host operating system's standard.
- **Trailing Slashes**: Trailing slashes are automatically stripped (e.g., `C:/Folder/` becomes `C:/Folder`), preventing errors in suffix matching or folder joins.

### Comments Support
This utility supports **JSON comments**. You are free to document your configurations directly inside JSON files using:
- **Single-line comments** starting with `//`
- **Multi-line block comments** starting with `/* ... */`

They will be stripped out before parsing, allowing for rich configuration notes.

### Structure Breakdown

The configuration file is written in standard JSON format (with comment support). It begins with a top-level key named `"dp"`, which groups all predefined properties. Inside `"dp"`, there is a `"paths"` section that contains `"src"` and `"dst"` fields specifying the source and destination paths:

1. **`src` (Object)**:
   - **`debug` (Array of Strings)**: Defines one or more absolute source directories where the **Debug build** output files are placed.
   - **`release` (Array of Strings)**: Defines one or more absolute source directories where the **Release build** output files are placed.
   - *Note: Source directory paths in both arrays must end with either `Win32` or `x64` (case-insensitive) so that the utility knows which architecture target files to copy.*

2. **`dst` (Object)**:
   - **`win32` (String)**: Specifies a custom destination folder for Win32 files. If left empty (`""`), omitted, or null, the utility automatically falls back to the default system-wide installation folder: `C:/Program Files (x86)/DigitalPersona/Bin` (or dynamically resolved via the `%ProgramFiles(x86)%` environment variable).
   - **`x64` (String)**: Specifies a custom destination folder for x64 files. If left empty (`""`), omitted, or null, the utility automatically falls back to the default system-wide installation folder: `C:/Program Files/DigitalPersona/Bin` (or dynamically resolved via the `%ProgramFiles%` environment variable).

### JSON Schema Template (with comments & mixed slashes example)

```json
{
  // Main DigitalPersona Configuration
  "dp": {
    "paths": {
      "src": {
        // Source folders (trailing slashes are automatically resolved and normalized)
        "debug": [
          "C:/y/c/dp/pm-native/src/~Output/Debug.Win32/",
          "C:/y/c/dp/pm-native/src/~Output/Debug.x64"
        ],
        "release": [
          "C:/y/c/dp/pm-native/src/~Output/Release.Win32",
          "C:/y/c/dp/pm-native/src/~Output/Release.x64/"
        ]
      },
      "dst": {
        /*
           Custom destinations:
           - Leave as "" (empty string) to automatically target default system folders:
             - win32: C:/Program Files (x86)/DigitalPersona/Bin
             - x64:   C:/Program Files/DigitalPersona/Bin
           - Provide a custom non-empty path (e.g. "C:/MyCustomFolder/Bin") to override.
        */
        "win32": "",
        "x64": ""
      }
    }
  }
}
```

*Note: Using forward slashes `/` in paths is highly recommended so that no escaping is required in JSON strings.*

---

## Example Configurations (`tests/` directory)

We have provided several configurations inside the `tests/` folder for reference or testing:

- **`config_full_dual_arch.json`**: Complete setup with both Win32 and x64 directories configured for both Debug and Release environments.
- **`config_x64_only.json`**: Restricts actions only to the 64-bit destination.
- **`config_win32_only.json`**: Restricts actions only to the 32-bit (x86) destination.
- **`config_custom_drive_paths.json`**: Demonstrates the use of alternate drives and directory naming layouts (e.g. `D:/BuildServer`).
- **`config_empty.json`**: Empty arrays structure, triggering registry fallbacks when run.

You can try using any of these by passing the `-config` flag:
```bash
# Run using the custom drive configuration example
npm start -- -config tests/config_custom_drive_paths.json
```
