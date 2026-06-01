package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadDefaults(t *testing.T) {
	os.Unsetenv("DUCT_SCRIPTS_DIR")
	os.Unsetenv("DUCT_TMP")

	cfg := Load("Ductfile", false)

	assert.Equal(t, "Ductfile", cfg.DuctfilePath)
	assert.False(t, cfg.LocalMode)
	assert.False(t, cfg.Debug)
	assert.NotEmpty(t, cfg.TmpDir)
}

func TestLoadLocalMode(t *testing.T) {
	cfg := Load("Ductfile", true)
	assert.True(t, cfg.LocalMode)
}

func TestLoadDebug(t *testing.T) {
	os.Setenv("DUCT_DEBUG", "true")
	defer os.Unsetenv("DUCT_DEBUG")

	cfg := Load("Ductfile", false)
	assert.True(t, cfg.Debug)
}

func TestLoadCustomScriptsDir(t *testing.T) {
	os.Setenv("DUCT_SCRIPTS_DIR", "/custom/scripts")
	defer os.Unsetenv("DUCT_SCRIPTS_DIR")

	cfg := Load("Ductfile", false)
	assert.Equal(t, "/custom/scripts", cfg.ScriptsDir)
}

func TestLoadCustomTmpDir(t *testing.T) {
	os.Setenv("DUCT_TMP", "/custom/tmp")
	defer os.Unsetenv("DUCT_TMP")

	cfg := Load("Ductfile", false)
	assert.Equal(t, "/custom/tmp", cfg.TmpDir)
}

func TestLoadScriptsDirFallback(t *testing.T) {
	os.Unsetenv("DUCT_SCRIPTS_DIR")

	// Create temp scripts dir
	tmpDir := t.TempDir()
	scriptsDir := filepath.Join(tmpDir, "scripts")
	os.MkdirAll(scriptsDir, 0755)

	// Change to temp dir
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cfg := Load("Ductfile", false)
	assert.Equal(t, scriptsDir, cfg.ScriptsDir)
}
