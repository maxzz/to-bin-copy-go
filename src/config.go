package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

type Config struct {
	Dp DpConfig `json:"dp"`
}

type DpConfig struct {
	Paths PathsConfig `json:"paths"`
}

type PathsConfig struct {
	Src SrcConfig `json:"src"`
	Dst DstConfig `json:"dst"`
}

type SrcConfig struct {
	Debug   []string `json:"debug"`
	Release []string `json:"release"`
}

type DstConfig struct {
	Win32 string `json:"win32"`
	X64   string `json:"x64"`
}

type AppArgs struct {
	IsRelease  bool
	ConfigPath string
	Source     string
}

func parseArgs() AppArgs {
	var args AppArgs
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	fs.Usage = func() {} // Override default usage behavior, we will print help ourselves

	fs.BoolVar(&args.IsRelease, "release", false, "Run in Release mode (default is Debug mode)")
	fs.StringVar(&args.ConfigPath, "config", "config.json", "Path to the configuration JSON file")
	fs.StringVar(&args.Source, "source", "", "Comma-separated list of custom source directories (bypasses config file/registry)")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			PrintHelp()
			os.Exit(0)
		}
		// Clean up and print the invalid argument error in Yellow
		errMsg := strings.TrimSpace(buf.String())
		fmt.Fprintln(os.Stderr, ColorYellow+"Error: "+errMsg+ColorReset)
		fmt.Println()
		PrintHelp()
		os.Exit(2)
	}

	// If there are any unexpected positional arguments, treat it as an invalid call.
	if fs.NArg() > 0 {
		fmt.Fprintln(os.Stderr, ColorYellow+fmt.Sprintf("Error: unexpected positional argument(s): %v", fs.Args())+ColorReset)
		fmt.Println()
		PrintHelp()
		os.Exit(1)
	}

	return args
}

// stripComments removes JavaScript/C++ style comments (//... and /*...*/) from a JSON byte slice,
// while correctly preserving characters inside string literals (even with escaped quotes).
func stripComments(data []byte) []byte {
	var result []byte
	inString := false
	inLineComment := false
	inBlockComment := false

	i := 0
	n := len(data)
	for i < n {
		if inLineComment {
			if data[i] == '\n' {
				inLineComment = false
				result = append(result, '\n')
			}
			i++
			continue
		}
		if inBlockComment {
			if i+1 < n && data[i] == '*' && data[i+1] == '/' {
				inBlockComment = false
				i += 2
			} else {
				i++
			}
			continue
		}

		if inString {
			if data[i] == '"' {
				escaped := false
				for j := i - 1; j >= 0; j-- {
					if data[j] == '\\' {
						escaped = !escaped
					} else {
						break
					}
				}
				if !escaped {
					inString = false
				}
			}
			result = append(result, data[i])
			i++
			continue
		}

		if i+1 < n && data[i] == '/' && data[i+1] == '/' {
			inLineComment = true
			i += 2
			continue
		}
		if i+1 < n && data[i] == '/' && data[i+1] == '*' {
			inBlockComment = true
			i += 2
			continue
		}
		if data[i] == '"' {
			inString = true
		}
		result = append(result, data[i])
		i++
	}
	return result
}

// loadConfig loads configuration from a JSON file, stripping comments first.
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cleanedData := stripComments(data)

	var cfg Config
	if err := json.Unmarshal(cleanedData, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// readRegistryConfig tries to read the configuration from Windows Registry.
func readRegistryConfig() (*Config, error) {
	const regPath = `SOFTWARE\SergeiMenchenin\CopyPmFilesToBin`
	k, err := registry.OpenKey(registry.CURRENT_USER, regPath, registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer k.Close()

	cfg := &Config{}
	
	debugPaths, _, err := k.GetStringsValue("sourcePathsDebug")
	if err == nil {
		cfg.Dp.Paths.Src.Debug = debugPaths
	}
	
	releasePaths, _, err := k.GetStringsValue("sourcePathsRelease")
	if err == nil {
		cfg.Dp.Paths.Src.Release = releasePaths
	}

	return cfg, nil
}

// splitUnescaped splits a string s by an unescaped separator sep (e.g. comma),
// using escape (e.g. backslash) to allow literal separators inside elements.
// Any occurrence of `\<sep>` (e.g. `\,`) is converted into `<sep>`.
// Other backslashes (such as standard Windows path backslashes like `\D`) are fully preserved.
func splitUnescaped(s string, sep rune, escape rune) []string {
	var segments []string
	var current strings.Builder
	runes := []rune(s)
	inEscape := false

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if inEscape {
			if r == sep || r == escape {
				current.WriteRune(r)
			} else {
				current.WriteRune(escape)
				current.WriteRune(r)
			}
			inEscape = false
		} else if r == escape {
			inEscape = true
		} else if r == sep {
			segments = append(segments, current.String())
			current.Reset()
		} else {
			current.WriteRune(r)
		}
	}
	if inEscape {
		current.WriteRune(escape)
	}
	segments = append(segments, current.String())
	return segments
}

// cleanAndNormalizePaths converts both backslashes and forward slashes to target OS separators,
// and removes any trailing slashes or redundant directory items.
func cleanAndNormalizePaths(paths []string) []string {
	normalized := make([]string, len(paths))
	for i, p := range paths {
		normalized[i] = filepath.Clean(filepath.FromSlash(strings.TrimSpace(p)))
	}
	return normalized
}

// GetSourcePaths resolves which source paths and destination paths to use.
// It returns (paths, dstConfig, modeName, sourceUsed, err).
func GetSourcePaths(args AppArgs) ([]string, DstConfig, string, string, error) {
	mode := "Debug"
	if args.IsRelease {
		mode = "Release"
	}

	var emptyDst DstConfig

	// 1. If explicit sources are given via CLI
	if args.Source != "" {
		paths := splitUnescaped(args.Source, ',', '\\')
		return cleanAndNormalizePaths(paths), emptyDst, mode, "CLI Flag", nil
	}

	// Helper to extract the right paths from Config
	getPathsFromConfig := func(cfg *Config) []string {
		if args.IsRelease {
			return cfg.Dp.Paths.Src.Release
		}
		return cfg.Dp.Paths.Src.Debug
	}

	// Helper to normalize the custom destinations if specified in config
	normalizeDst := func(dst DstConfig) DstConfig {
		if dst.Win32 != "" {
			dst.Win32 = filepath.Clean(filepath.FromSlash(dst.Win32))
		}
		if dst.X64 != "" {
			dst.X64 = filepath.Clean(filepath.FromSlash(dst.X64))
		}
		return dst
	}

	// 2. Try loading config from specified config path (or default "config.json")
	cfg, err := loadConfig(args.ConfigPath)
	if err == nil {
		paths := getPathsFromConfig(cfg)
		if len(paths) > 0 {
			return cleanAndNormalizePaths(paths), normalizeDst(cfg.Dp.Paths.Dst), mode, fmt.Sprintf("Config File (%s)", args.ConfigPath), nil
		}
	}

	// 2b. If the specified config was "config.json", also try next to the executable
	if args.ConfigPath == "config.json" {
		if exePath, err := os.Executable(); err == nil {
			exeDir := filepath.Dir(exePath)
			altConfigPath := filepath.Join(exeDir, "config.json")
			if altConfigPath != args.ConfigPath {
				if cfg, err := loadConfig(altConfigPath); err == nil {
					paths := getPathsFromConfig(cfg)
					if len(paths) > 0 {
						return cleanAndNormalizePaths(paths), normalizeDst(cfg.Dp.Paths.Dst), mode, fmt.Sprintf("Config File (%s)", altConfigPath), nil
					}
				}
			}
		}
	}

	// 3. Fallback to Windows Registry
	regCfg, err := readRegistryConfig()
	if err == nil {
		paths := getPathsFromConfig(regCfg)
		if len(paths) > 0 {
			return cleanAndNormalizePaths(paths), normalizeDst(regCfg.Dp.Paths.Dst), mode, "Windows Registry", nil
		}
	}

	// 4. No paths found
	return nil, emptyDst, mode, "", fmt.Errorf("no source paths found! Please configure them in config.json, Windows Registry, or pass via the -source flag")
}
