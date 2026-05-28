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
	Items []ConfigItem `json:"items"`
}

type ConfigItem struct {
	Name     string      `json:"name"`
	IsActive bool        `json:"isActive"`
	Paths    *PathBlock  `json:"paths,omitempty"`
	Files    []FileBlock `json:"files,omitempty"`
	Wait     *bool       `json:"wait,omitempty"`
}

type PathBlock struct {
	Dp              bool      `json:"dp"`
	Src             SrcConfig `json:"src"`
	Dst             DstConfig `json:"dst"`
	SrcFilesInclude []string  `json:"srcFilesInclude,omitempty"`
	SrcFilesExclude []string  `json:"srcFilesExclude,omitempty"`
	Restart         *bool     `json:"restart,omitempty"`
}

type SrcConfig struct {
	Debug   []string `json:"debug"`
	Release []string `json:"release"`
}

type DstConfig struct {
	Win32 string `json:"win32"`
	X64   string `json:"x64"`
}

type FileBlock struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}

type AppArgs struct {
	IsRelease     bool
	ConfigPath    string
	Source        string
	SetName       string
	Force         bool
	Restart       bool
	Wait          bool
	WaitSpecified bool
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
	fs.StringVar(&args.SetName, "set", "", "Name of a specific configuration set to execute")
	fs.BoolVar(&args.Force, "force", false, "Force copy all files regardless of timestamps")
	fs.BoolVar(&args.Restart, "restart", false, "Automatically restart DPAgent.exe immediately after all files have finished copying (if dp is true)")
	fs.BoolVar(&args.Wait, "wait", false, "Wait and display success/error screens after execution is completed")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			PrintHelp()
			os.Exit(0)
		}
		// Clean up and print the invalid argument error in Yellow
		errMsg := strings.TrimSpace(buf.String())
		fmt.Fprintln(os.Stderr, ColorYellow+"Error: "+errMsg+ColorReset)
		fmt.Fprintln(os.Stderr, ColorYellow+"Run 'copy-pm-files.exe -help' to see full usage instructions."+ColorReset)
		os.Exit(2)
	}

	configSpecified := false
	sourceSpecified := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "config" {
			configSpecified = true
		}
		if f.Name == "source" {
			sourceSpecified = true
		}
		if f.Name == "wait" {
			args.WaitSpecified = true
		}
	})

	if !configSpecified && !sourceSpecified {
		fmt.Fprintln(os.Stderr, ColorYellow+"Warning: No configuration file was specified and the --source option remains undefined."+ColorReset)
		fmt.Println()
		fmt.Fprintln(os.Stderr, ColorDim+"Run 'copy-pm-files.exe -help' to see full usage instructions."+ColorReset)
		os.Exit(1)
	}

	// If there are any unexpected positional arguments, treat it as an invalid call.
	if fs.NArg() > 0 {
		fmt.Fprintln(os.Stderr, ColorYellow+fmt.Sprintf("Error: unexpected positional argument(s): %v", fs.Args())+ColorReset)
		fmt.Println()
		fmt.Fprintln(os.Stderr, ColorDim+"Run 'copy-pm-files.exe -help' to see full usage instructions."+ColorReset)
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
func readRegistryConfig() (*PathBlock, error) {
	const regPath = `SOFTWARE\AATanam\CopyPmFilesToBin`
	k, err := registry.OpenKey(registry.CURRENT_USER, regPath, registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer k.Close()

	pb := &PathBlock{Dp: true}

	debugPaths, _, err := k.GetStringsValue("sourcePathsDebug")
	if err == nil {
		pb.Src.Debug = debugPaths
	}

	releasePaths, _, err := k.GetStringsValue("sourcePathsRelease")
	if err == nil {
		pb.Src.Release = releasePaths
	}

	return pb, nil
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

// ResolveConfigAndItems resolves the selected items to process and returns (items, configFilePath, error).
func ResolveConfigAndItems(args AppArgs) ([]ConfigItem, string, error) {
	// 1. If explicit sources are given via CLI
	if args.Source != "" {
		paths := splitUnescaped(args.Source, ',', '\\')
		cleaned := cleanAndNormalizePaths(paths)
		item := ConfigItem{
			Name:     "CLI Source Flag Override",
			IsActive: true,
			Paths: &PathBlock{
				Dp: true,
				Src: SrcConfig{
					Debug:   cleaned,
					Release: cleaned,
				},
			},
		}
		return []ConfigItem{item}, "", nil
	}

	// 2. Try loading config from specified config path (or default "config.json")
	cfg, err := loadConfig(args.ConfigPath)
	var loadedPath string
	if err == nil {
		loadedPath, _ = filepath.Abs(args.ConfigPath)
	} else if args.ConfigPath == "config.json" {
		// 2b. Try next to the executable if default "config.json" failed to load
		if exePath, errExe := os.Executable(); errExe == nil {
			exeDir := filepath.Dir(exePath)
			altConfigPath := filepath.Join(exeDir, "config.json")
			if altConfigPath != args.ConfigPath {
				cfg, errExe = loadConfig(altConfigPath)
				if errExe == nil {
					loadedPath, _ = filepath.Abs(altConfigPath)
					err = nil // clear error
				}
			}
		}
	}

	var items []ConfigItem
	if err == nil && cfg != nil && len(cfg.Items) > 0 {
		// Filter items based on SetName
		if args.SetName != "" {
			for _, item := range cfg.Items {
				if strings.EqualFold(item.Name, args.SetName) {
					items = append(items, item)
				}
			}
			if len(items) == 0 {
				return nil, "", fmt.Errorf("no configuration set found with name %q", args.SetName)
			}
		} else {
			// Select active items
			for _, item := range cfg.Items {
				if item.IsActive {
					items = append(items, item)
				}
			}
		}
	}

	if len(items) > 0 {
		return items, loadedPath, nil
	}

	// 3. Fallback to Windows Registry
	regBlock, regErr := readRegistryConfig()
	if regErr == nil && regBlock != nil {
		// Verify if registry has any paths
		if len(regBlock.Src.Debug) > 0 || len(regBlock.Src.Release) > 0 {
			item := ConfigItem{
				Name:     "Windows Registry Fallback",
				IsActive: true,
				Paths:    regBlock,
			}
			return []ConfigItem{item}, "", nil
		}
	}

	// Compile a meaningful error message
	var errMsg string
	if err != nil {
		errMsg = fmt.Sprintf("failed to load config file: %v. Registry fallback also failed", err)
	} else {
		errMsg = "no source paths or active items found in config file. Registry fallback also failed"
	}
	return nil, "", fmt.Errorf("%s", errMsg)
}
