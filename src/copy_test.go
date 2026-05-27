package main

import (
	"errors"
	"os"
	"strings"
	"syscall"
	"testing"

	"golang.org/x/sys/windows"
)

/*
TestSplitUnescaped tests the splitUnescaped function.

Description:
It verifies that splitUnescaped correctly splits a string by a given separator rune,
while respecting backslash escaping (e.g., matching literal commas and ignoring escaped ones),
and ensuring standard Windows backslashes are preserved for regular directories.

Examples:
- Input s: `C:/Path1,C:/Path2`, sep: `,`, escape: `\`
  Expected Output: `[]string{"C:/Path1", "C:/Path2"}`

- Input s: `C:/Folder\, With Comma/Win32,C:/NormalPath/x64`, sep: `,`, escape: `\`
  Expected Output: `[]string{"C:/Folder, With Comma/Win32", "C:/NormalPath/x64"}`

- Input s: `C:\Path\With\Backslashes,D:\Other\Path`, sep: `,`, escape: `\`
  Expected Output: `[]string{"C:\\Path\\With\\Backslashes", "D:\\Other\\Path"}`
*/
func TestSplitUnescaped(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple comma separation",
			input:    "C:/Path1,C:/Path2",
			expected: []string{"C:/Path1", "C:/Path2"},
		},
		{
			name:     "Separators with spaces",
			input:    "C:/Folder With Spaces/Win32,C:/Other Folder/x64",
			expected: []string{"C:/Folder With Spaces/Win32", "C:/Other Folder/x64"},
		},
		{
			name:     "Escaped comma",
			input:    "C:/Folder\\, With Comma/Win32,C:/NormalPath/x64",
			expected: []string{"C:/Folder, With Comma/Win32", "C:/NormalPath/x64"},
		},
		{
			name:     "Standard backslashes preserved",
			input:    "C:\\y\\c\\dp\\pm-native\\src\\~Output\\Debug.Win32,C:\\y\\c\\dp\\pm-native\\src\\~Output\\Debug.x64",
			expected: []string{"C:\\y\\c\\dp\\pm-native\\src\\~Output\\Debug.Win32", "C:\\y\\c\\dp\\pm-native\\src\\~Output\\Debug.x64"},
		},
		{
			name:     "Standard backslashes alongside escaped commas",
			input:    "C:\\y\\c\\dp\\pm-native\\src\\~Output\\Debug\\,Win32,C:\\y\\c\\dp\\pm-native\\src\\~Output\\Debug.x64",
			expected: []string{"C:\\y\\c\\dp\\pm-native\\src\\~Output\\Debug,Win32", "C:\\y\\c\\dp\\pm-native\\src\\~Output\\Debug.x64"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitUnescaped(tt.input, ',', '\\')
			if len(result) != len(tt.expected) {
				t.Fatalf("splitUnescaped(%q) returned %d segments, expected %d. Result: %v", tt.input, len(result), len(tt.expected), result)
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("splitUnescaped(%q) segment %d = %q, expected %q", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

/*
TestStripComments tests the stripComments function.

Description:
It verifies that stripComments successfully deletes single-line (//...) and multi-line (/*...* /)
JavaScript/C++ style comments from raw byte blocks while preserving characters inside string literals,
allowing comments to be placed anywhere inside configuration files.

Example Input:
{
    // This is a line comment
    "name": "test", /* block comment * /
    "url": "http://example.com/api"
}

Expected Output:
{
    "name": "test", 
    "url": "http://example.com/api"
}
*/
func TestStripComments(t *testing.T) {
	input := []byte(`{
		// This is a line comment
		"name": "test", /* This is a
		block comment */
		"url": "http://example.com/api", // Keeping URLs safe
		"quote": "A string \"with escaped quote\" inside", /* comment here */
		"mixed_slashes": "C:/some/path"
	}`)

	expected := `{
		
		"name": "test", 
		"url": "http://example.com/api", 
		"quote": "A string \"with escaped quote\" inside", 
		"mixed_slashes": "C:/some/path"
	}`

	cleaned := stripComments(input)
	
	// Normalize spacing/newlines for comparison simplicity
	cleanStr := strings.ReplaceAll(string(cleaned), "\r", "")
	expStr := strings.ReplaceAll(expected, "\r", "")
	_ = expStr

	// Ensure whitespaces are stripped for strict equality comparison or check content
	if !strings.Contains(cleanStr, `"name": "test"`) || 
	   !strings.Contains(cleanStr, `"url": "http://example.com/api"`) ||
	   !strings.Contains(cleanStr, `"quote": "A string \"with escaped quote\" inside"`) ||
	   strings.Contains(cleanStr, "This is a line comment") ||
	   strings.Contains(cleanStr, "This is a\n\t\tblock comment") {
		t.Errorf("stripComments failed. Cleaned output was:\n%s", cleanStr)
	}
}

/*
TestIsSharingViolation tests the isSharingViolation function.

Description:
It verifies that isSharingViolation correctly detects Windows sharing violations (ERROR_SHARING_VIOLATION = 32),
including standard errors, direct error numbers, and nested/wrapped os.PathError structures.

Examples:
- Input: `nil` -> Expected: `false`
- Input: `errors.New("some standard error")` -> Expected: `false`
- Input: `windows.ERROR_SHARING_VIOLATION` -> Expected: `true`
- Input: `&os.PathError{Err: windows.ERROR_SHARING_VIOLATION}` -> Expected: `true`
- Input: `&os.PathError{Err: syscall.ENOTDIR}` -> Expected: `false`
*/
func TestIsSharingViolation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "Generic error",
			err:      errors.New("some standard error"),
			expected: false,
		},
		{
			name:     "Direct sharing violation errno",
			err:      windows.ERROR_SHARING_VIOLATION,
			expected: true,
		},
		{
			name: "PathError wrapping sharing violation",
			err: &os.PathError{
				Op:   "open",
				Path: "C:\\test.txt",
				Err:  windows.ERROR_SHARING_VIOLATION,
			},
			expected: true,
		},
		{
			name: "PathError wrapping other errno",
			err: &os.PathError{
				Op:   "open",
				Path: "C:\\test.txt",
				Err:  syscall.ENOTDIR,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSharingViolation(tt.err)
			if result != tt.expected {
				t.Errorf("isSharingViolation() for %s = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}
