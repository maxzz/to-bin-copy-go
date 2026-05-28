package main

import (
	"os"
	"path/filepath"
	"testing"
)

/*
TestCleanAndNormalizePaths tests the cleanAndNormalizePaths function.

Description:
It verifies that a slice of raw file paths containing forward slashes, backslashes, mixed
slashes, trailing slashes, and redundant spaces is correctly cleaned and normalized to
conform to the host operating system's standard file path format.

Examples:

  - Input:  []string{"C:/y/c/dp/  ", "  D:\\some\\other/path/ "}
    Output: []string{"C:\\y\\c\\dp", "D:\\some\\other\\path"} (on Windows)

  - Input:  []string{"C:/y/c/dp/", "D:\\some\\path\\"}
    Output: []string{"C:\\y\\c\\dp", "D:\\some\\path"} (on Windows)
*/
func TestCleanAndNormalizePaths(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Mixed slashes and spaces",
			input:    []string{"C:/y/c/dp/  ", "  D:\\some\\other/path/ "},
			expected: []string{filepath.Clean("C:/y/c/dp"), filepath.Clean("D:/some/other/path")},
		},
		{
			name:     "Trailing slashes",
			input:    []string{"C:/y/c/dp/", "D:\\some\\path\\"},
			expected: []string{filepath.Clean("C:/y/c/dp"), filepath.Clean("D:/some/path")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanAndNormalizePaths(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d paths, got %d", len(tt.expected), len(result))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("at index %d: expected %q, got %q", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

/*
TestLoadConfig tests the loadConfig function.

Description:
It verifies that the loadConfig function successfully reads and parses the JSON configuration
structure. It also ensures that stripComments correctly filters out JavaScript-style single-line (//)
and multi-line (/* ... * /) comments from the file contents prior to JSON unmarshaling, without
breaking nested string fields or valid brackets.

Example JSON input with comments:

	{
	    // Comment line
	    "items": [
	        {
	            "name": "set-1",
	            "isActive": true,
	            "paths": { "dp": true, "src": { "debug": ["C:/src/Win32"] } }
	        }
	    ]
	}

Expected output parsed struct:

	Config{
	    Items: []ConfigItem{
	        {
	            Name: "set-1",
	            IsActive: true,
	            Paths: &PathBlock{ Dp: true, Src: SrcConfig{ Debug: ["C:/src/Win32"] } }
	        }
	    }
	}
*/
func TestLoadConfig(t *testing.T) {
	// Create a temporary config file with comments to ensure our stripComments works
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configContent := `{
		// A comment explaining the config
		"items": [
			{
				"name": "set-1",
				"isActive": true,
				"wait": true,
				"paths": {
					"dp": true,
					"restart": true,
					"src": {
						"debug": ["C:/src/Win32"],
						"release": ["C:/src/Release.x64"]
					},
					"dst": {
						"win32": "C:/dst/Win32",
						"x64": "C:/dst/x64"
					}
				}
			},
			/*
			   A block comment explaining another item
			*/
			{
				"name": "set-2",
				"isActive": false,
				"files": [
					{
						"src": "C:/src/file.sys",
						"dst": "C:/dst/file.sys"
					}
				]
			}
		]
	}`

	tmpFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(tmpFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}

	cfg, err := loadConfig(tmpFile)
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}

	if len(cfg.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(cfg.Items))
	}

	item1 := cfg.Items[0]
	if item1.Name != "set-1" || !item1.IsActive || item1.Paths == nil || !item1.Paths.Dp {
		t.Errorf("item 1 parsed incorrectly: %+v", item1)
	}

	if item1.Wait == nil || !*item1.Wait {
		t.Errorf("item 1 Wait expected to be true, got %+v", item1.Wait)
	}

	if len(item1.Paths.Src.Debug) != 1 || item1.Paths.Src.Debug[0] != "C:/src/Win32" {
		t.Errorf("item 1 Src.Debug parsed incorrectly")
	}

	if item1.Paths.Restart == nil || !*item1.Paths.Restart {
		t.Errorf("item 1 Restart expected to be true, got %+v", item1.Paths.Restart)
	}

	item2 := cfg.Items[1]
	if item2.Name != "set-2" || item2.IsActive || item2.Paths != nil || len(item2.Files) != 1 {
		t.Errorf("item 2 parsed incorrectly: %+v", item2)
	}

	if item2.Files[0].Src != "C:/src/file.sys" || item2.Files[0].Dst != "C:/dst/file.sys" {
		t.Errorf("item 2 Files parsed incorrectly")
	}
}

/*
TestResolveConfigAndItems tests the ResolveConfigAndItems function.

Description:
It verifies the hierarchical config resolution, flag priority, and filtering logic of our application:
 1. **Case 1 (CLI Override)**: Supplying a manual `-source` CLI flag bypasses file-loading entirely and
    generates an ad-hoc configuration item.
    - Example Input flags: `AppArgs{Source: "C:/custom/Win32,C:/custom/x64"}`
    - Expected Output: `[]ConfigItem` containing 1 item named "CLI Source Flag Override", `configPath = ""`
 2. **Case 2 (Active Sets)**: Running with a config file without specifying a `-set` flag filters and
    returns only items where `isActive` is `true`.
    - Example Input flags: `AppArgs{ConfigPath: tmpFile}`
    - Expected Output: Only returns `active-set`
 3. **Case 3 (Set Selection)**: Supplying a specific set name using the `-set` flag filters and returns
    only that matching item, completely ignoring its default `isActive` flag.
    - Example Input flags: `AppArgs{ConfigPath: tmpFile, SetName: "inactive-set"}`
    - Expected Output: Returns `inactive-set` (even though its isActive is false)
 4. **Case 4 (Non-Existent Set Error)**: Requesting a set name that does not exist in the file returns
    a clean and helpful error.
    - Example Input flags: `AppArgs{ConfigPath: tmpFile, SetName: "does-not-exist"}`
    - Expected Output: `err != nil`
*/
func TestResolveConfigAndItems(t *testing.T) {
	// Create a temporary config file
	tmpDir, err := os.MkdirTemp("", "resolve-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configContent := `{
		"items": [
			{
				"name": "active-set",
				"isActive": true,
				"paths": {
					"dp": true,
					"src": {
						"debug": ["C:/src/Win32"]
					}
				}
			},
			{
				"name": "inactive-set",
				"isActive": false,
				"paths": {
					"dp": false,
					"src": {
						"debug": ["C:/src/x64"]
					}
				}
			}
		]
	}`

	tmpFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(tmpFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}

	// Case 1: CLI override (-source specified)
	argsCLI := AppArgs{
		Source: "C:/custom/Win32,C:/custom/x64",
	}
	items, configPath, err := ResolveConfigAndItems(argsCLI)
	if err != nil {
		t.Fatalf("failed resolving CLI override: %v", err)
	}
	if len(items) != 1 || items[0].Name != "CLI Source Flag Override" {
		t.Errorf("expected CLI Source Flag Override, got %+v", items)
	}
	if configPath != "" {
		t.Errorf("expected empty configPath for CLI override, got %q", configPath)
	}

	// Case 2: Config file, no -set specified (should select active-set only)
	argsActiveOnly := AppArgs{
		ConfigPath: tmpFile,
	}
	items, configPath, err = ResolveConfigAndItems(argsActiveOnly)
	if err != nil {
		t.Fatalf("failed resolving active items: %v", err)
	}
	if len(items) != 1 || items[0].Name != "active-set" {
		t.Errorf("expected 'active-set' only, got %d items: %+v", len(items), items)
	}
	expectedAbsPath, _ := filepath.Abs(tmpFile)
	if configPath != expectedAbsPath {
		t.Errorf("expected configPath %q, got %q", expectedAbsPath, configPath)
	}

	// Case 3: Config file, -set inactive-set specified (should select inactive-set)
	argsWithSet := AppArgs{
		ConfigPath: tmpFile,
		SetName:    "inactive-set",
	}
	items, configPath, err = ResolveConfigAndItems(argsWithSet)
	if err != nil {
		t.Fatalf("failed resolving with -set flag: %v", err)
	}
	if len(items) != 1 || items[0].Name != "inactive-set" {
		t.Errorf("expected 'inactive-set', got %+v", items)
	}

	// Case 4: Config file, -set non-existent (should fail)
	argsNonExistentSet := AppArgs{
		ConfigPath: tmpFile,
		SetName:    "does-not-exist",
	}
	_, _, err = ResolveConfigAndItems(argsNonExistentSet)
	if err == nil {
		t.Error("expected failure for non-existent set name, got nil error")
	}
}
