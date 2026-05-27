package main

import (
	"errors"
	"os"
	"syscall"
	"testing"

	"golang.org/x/sys/windows"
)

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
