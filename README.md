# DigitalPersona PM File Copier (Go Version)

A robust, lightweight command-line utility written in Go for copying DigitalPersona PM native binaries and driver files from build output directories to the system's DigitalPersona installation directory.

This utility is a direct conversion of the original C# WinForms application `CopyPmFilesToBin` to Go. It keeps all of the original functionality intact, adding a cleaner command-line interface, detailed progress logs, and helpful system privilege warnings.

---

## Table of Contents

- [Key Features](#key-features)
- [Configuration File Format & Structure](#configuration-file-format--structure)
  - [Flexible Path Syntax (Slashes & Normalization)](#flexible-path-syntax-slashes--normalization)
  - [Comments Support](#comments-support)
  - [Structure Breakdown](#structure-breakdown)
  - [JSON Schema Template (with comments & mixed slashes example)](#json-schema-template-with-comments--mixed-slashes-example)
- [Example Configurations (tests/ directory)](#example-configurations-tests-directory)
- [Folder Structure](#folder-structure)
- [How to Build and Run (Using NPM/Node scripts)](#how-to-build-and-run-using-npm-node-scripts)
  - [Using NPM Scripts](#using-npm-scripts)
- [How to Build (Using Native Go Commands)](#how-to-build-using-native-go-commands)
- [How to Run](#how-to-run)
  - [Command-Line Arguments](#command-line-arguments)
  - [Examples](#examples)
- [Detailed Configuration & Target Mode Clarification](#detailed-configuration--target-mode-clarification)
  - [Important: Target Modes vs. Utility Compilation](#important-target-modes-vs-utility-compilation)

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

The configuration file is written in standard JSON format (with comment support). It begins with a top-level key named `"items"`, which is an array of configuration blocks/action sets.

Each item in `"items"` contains:
- **`name` (String, Optional)**: The unique identifier for this configuration set. If specified, it can be selected via the `-set <name>` flag on startup.
- **`isActive` (Boolean)**: If set to `true`, this set will be processed during a default run (when no specific set is designated using the `-set` flag).
- **`paths` (Object, Optional)**: Contains details on directory paths and inclusion/exclusion matching:
  - **`dp` (Boolean)**: If `true`, the application executes its standard predefined file copies (such as `win32Files` or `x64Files`). If `false` or omitted, custom folder-to-folder file copying is performed based on directories.
  - **`src` (Object)**:
    - **`debug` (Array of Strings)**: Source directories where the Debug compiled files reside.
    - **`release` (Array of Strings)**: Source directories where the Release compiled files reside.
    - *Note: Source directories must end with either `Win32` or `x64` to designate their architecture.*
  - **`dst` (Object)**:
    - **`win32` (String)**: Target folder for Win32 files (defaults to `%ProgramFiles(x86)%\DigitalPersona\Bin`).
    - **`x64` (String)**: Target folder for x64 files (defaults to `%ProgramFiles%\DigitalPersona\Bin`).
  - **`srcFilesInclude` (Array of Strings, Optional)**: Active only when `dp` is `false`. List of regular expression patterns. Only files whose relative paths from the source directory match any of these patterns are copied. If omitted, all files are matched by default.
  - **`srcFilesExclude` (Array of Strings, Optional)**: Active only when `dp` is `false`. List of regular expression patterns. Any files matching these patterns are excluded from being copied.
- **`files` (Array of Objects, Optional)**: An alternative block of actions specifying individual files to copy directly:
  - **`src` (String)**: Full path and filename of the source file.
  - **`dst` (String)**: Full path and filename of the destination file.

### JSON Schema Template (with comments & mixed slashes example)

```json
{
  "items": [
    {
      "name": "dp-binaries",
      "isActive": true,
      "paths": {
        "dp": true,
        "src": {
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
          "win32": "",
          "x64": ""
        }
      }
    },
    {
      "name": "custom-wildcard-copier",
      "isActive": false,
      "paths": {
        "dp": false,
        "src": {
          "debug": [
            "C:/build/Debug.Win32"
          ],
          "release": [
            "C:/build/Release.Win32"
          ]
        },
        "dst": {
          "win32": "C:/Target/Win32"
        },
        "srcFilesInclude": [
          "\\.dll$",
          "\\.exe$"
        ],
        "srcFilesExclude": [
          "test.*"
        ]
      }
    },
    {
      "name": "individual-files",
      "isActive": true,
      "files": [
        {
          "src": "C:/my_source/custom_driver.sys",
          "dst": "C:/Windows/System32/drivers/custom_driver.sys"
        }
      ]
    }
  ]
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

---

## Folder Structure

- `src/`: Directory containing all Go source files.
  - `main.go`: Entry point and program orchestration.
  - `config.go`: Argument parsing, JSON configuration, and Windows Registry integration.
  - `help.go`: Modular help output displaying usage, default options, and behaviors.
  - `process.go`: Graceful Win32 process search and close logic for `DPAgent.exe`.
  - `copy.go`: Core copy loop, timestamp comparison, and locked file renaming.
  - `copy_test.go`: Unit tests for file utility behaviors.
- `scripts/`: Development and utility scripts.
  - `build.js`: Auto-incrementing version builder that compiles the Go binary.
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

- `-release`: Run the utility targeting **Release** compiled files. By default, the utility operates on **Debug** compiled files.
- `-config <path>`: Path to a JSON configuration file (defaults to `config.json`).
- `-source <paths>`: Comma-separated list of paths (e.g. `-source "C:/src/Win32,C:/src/x64"`), bypassing any registry or config file settings.
  - **Handling Spaces**: If any paths contain spaces, you **must enclose the entire comma-separated list of paths in double quotes**. For example:
    `.\copy-pm-files.exe -source "C:/Folder With Spaces/Win32,C:/Other Folder/x64"`
  - **Handling Commas**: If a directory or file name contains a literal comma (`,`), you **must escape the comma with a backslash (`\,`)** inside the list. For example:
    `.\copy-pm-files.exe -source "C:/Folder\, With Comma/Win32,C:/NormalFolder/x64"`
- `-set <name>`: Optional. Designate a specific configuration set from `config.json` to execute (and only this one)—regardless of whether its `isActive` attribute is set to `true` or `false`.
- `-force`: Optional. Force copy all files regardless of timestamps ("copy if newer" is the default).

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
