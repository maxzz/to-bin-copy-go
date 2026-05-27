package main

import (
	"os"
	"path/filepath"
	"testing"
)

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
				"paths": {
					"dp": true,
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

	if len(item1.Paths.Src.Debug) != 1 || item1.Paths.Src.Debug[0] != "C:/src/Win32" {
		t.Errorf("item 1 Src.Debug parsed incorrectly")
	}

	item2 := cfg.Items[1]
	if item2.Name != "set-2" || item2.IsActive || item2.Paths != nil || len(item2.Files) != 1 {
		t.Errorf("item 2 parsed incorrectly: %+v", item2)
	}

	if item2.Files[0].Src != "C:/src/file.sys" || item2.Files[0].Dst != "C:/dst/file.sys" {
		t.Errorf("item 2 Files parsed incorrectly")
	}
}

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
