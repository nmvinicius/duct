package config

import (
	"os"
	"path/filepath"
	"strconv"
)

// Config holds runtime configuration
type Config struct {
	DuctfilePath string
	LocalMode    bool
	Debug        bool
	ScriptsDir   string
	TmpDir       string
}

// Load creates a new Config from environment and flags
func Load(ductfilePath string, localMode bool) *Config {
	scriptsDir := os.Getenv("DUCT_SCRIPTS_DIR")
	if scriptsDir == "" {
		// Try to find scripts relative to binary or current dir
		if _, err := os.Stat("./scripts"); err == nil {
			scriptsDir, _ = filepath.Abs("./scripts")
		}
	}

	tmpDir := os.Getenv("DUCT_TMP")
	if tmpDir == "" {
		tmpDir = "/tmp/duct-" + strconv.Itoa(os.Getpid())
	}

	return &Config{
		DuctfilePath: ductfilePath,
		LocalMode:    localMode,
		Debug:        os.Getenv("DUCT_DEBUG") == "true",
		ScriptsDir:   scriptsDir,
		TmpDir:       tmpDir,
	}
}
