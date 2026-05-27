package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

type Config struct {
	SourcePathsDebug   []string `json:"sourcePathsDebug"`
	SourcePathsRelease []string `json:"sourcePathsRelease"`
}

type AppArgs struct {
	IsRelease  bool
	ConfigPath string
	Source     string
}

func parseArgs() AppArgs {
	var args AppArgs
	flag.BoolVar(&args.IsRelease, "release", false, "Run in Release mode (default is Debug mode)")
	flag.StringVar(&args.ConfigPath, "config", "config.json", "Path to the configuration JSON file")
	flag.StringVar(&args.Source, "source", "", "Comma-separated list of custom source directories (bypasses config file/registry)")
	flag.Parse()
	return args
}

// loadConfig loads configuration from a JSON file.
func loadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	dec := json.NewDecoder(file)
	if err := dec.Decode(&cfg); err != nil {
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
		cfg.SourcePathsDebug = debugPaths
	}
	
	releasePaths, _, err := k.GetStringsValue("sourcePathsRelease")
	if err == nil {
		cfg.SourcePathsRelease = releasePaths
	}

	return cfg, nil
}

// GetSourcePaths resolves which source paths to use.
// It returns (paths, modeName, sourceUsed, err).
func GetSourcePaths(args AppArgs) ([]string, string, string, error) {
	mode := "Debug"
	if args.IsRelease {
		mode = "Release"
	}

	// 1. If explicit sources are given via CLI
	if args.Source != "" {
		paths := strings.Split(args.Source, ",")
		for i, p := range paths {
			paths[i] = strings.TrimSpace(p)
		}
		return paths, mode, "CLI Flag", nil
	}

	// Helper to extract the right paths from Config
	getPathsFromConfig := func(cfg *Config) []string {
		if args.IsRelease {
			return cfg.SourcePathsRelease
		}
		return cfg.SourcePathsDebug
	}

	// 2. Try loading config from specified config path (or default "config.json")
	cfg, err := loadConfig(args.ConfigPath)
	if err == nil {
		paths := getPathsFromConfig(cfg)
		if len(paths) > 0 {
			return paths, mode, fmt.Sprintf("Config File (%s)", args.ConfigPath), nil
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
						return paths, mode, fmt.Sprintf("Config File (%s)", altConfigPath), nil
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
			return paths, mode, "Windows Registry", nil
		}
	}

	// 4. No paths found
	return nil, mode, "", fmt.Errorf("no source paths found! Please configure them in config.json, Windows Registry, or pass via the -source flag")
}
