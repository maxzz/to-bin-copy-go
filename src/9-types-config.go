package main

// Config represents the top-level configuration structure.
type Config struct {
	Items []ConfigItem `json:"items"`
}

// ConfigItem represents an individual configuration set.
type ConfigItem struct {
	SetName  string      `json:"setName"`
	IsActive bool        `json:"isActive"`
	Paths    *PathBlock  `json:"paths,omitempty"`
	Files    []FileBlock `json:"files,omitempty"`
	Wait     *bool       `json:"wait,omitempty"`
}

// PathBlock defines the source and destination paths along with optional behaviors.
type PathBlock struct {
	Dp              bool      `json:"dp"`
	Src             SrcConfig `json:"src"`
	Dst             DstConfig `json:"dst"`
	SrcFilesInclude []string  `json:"srcFilesInclude,omitempty"`
	SrcFilesExclude []string  `json:"srcFilesExclude,omitempty"`
	Restart         *bool     `json:"restart,omitempty"`
}

// SrcConfig holds source directory paths for both Debug and Release environments.
type SrcConfig struct {
	Debug   []string `json:"debug"`
	Release []string `json:"release"`
}

// DstConfig specifies standard target destinations.
type DstConfig struct {
	Win32 string `json:"win32"`
	X64   string `json:"x64"`
}

// FileBlock defines direct file copying pairs.
type FileBlock struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}
