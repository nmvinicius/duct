package executor

import (
	"strings"
	"testing"

	"github.com/nmvinicius/duct/internal/config"
	"github.com/nmvinicius/duct/internal/platform"
	"github.com/nmvinicius/duct/pkg/ductfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExecutor(t *testing.T) {
	df := &ductfile.Ductfile{
		Project: "test",
		Globals: map[string]string{
			"FOO": "bar",
		},
	}

	cfg := &config.Config{
		ScriptsDir: "/tmp/scripts",
		TmpDir:     "/tmp/duct",
	}

	exec := New(df, platform.Local, cfg)

	assert.Equal(t, "test", exec.env["PROJECT"])
	assert.Equal(t, "bar", exec.env["FOO"])
	assert.Equal(t, "0.1.0", exec.env["DUCT_VERSION"])
}

func TestExpandVars(t *testing.T) {
	exec := &Executor{
		env: map[string]string{
			"PROJECT": "myapp",
			"TAG":     "v1",
		},
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"$PROJECT", "myapp"},
		{"${PROJECT}", "myapp"},
		{"image:$TAG", "image:v1"},
		{"$PROJECT:$TAG", "myapp:v1"},
		{"no vars here", "no vars here"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := exec.expandVars(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestDryRun(t *testing.T) {
	df := &ductfile.Ductfile{
		Project: "test",
		Steps: []ductfile.Step{
			{
				Name: "build",
				Uses: []string{"node"},
				Runs: []string{"echo $PROJECT"},
			},
			{
				Name:  "deploy",
				Needs: []string{"build"},
				When: &ductfile.Condition{
					Raw:    `branch == "main"`,
					Branch: `== "main"`,
				},
				Runs: []string{"echo deploy"},
			},
		},
	}

	cfg := &config.Config{}
	err := DryRun(df, platform.Local, cfg)
	require.NoError(t, err)
}

func TestExecutorRunStepNotFound(t *testing.T) {
	df := &ductfile.Ductfile{
		Steps: []ductfile.Step{
			{Name: "build", Runs: []string{"echo ok"}},
		},
	}

	exec := New(df, platform.Local, &config.Config{})
	err := exec.RunStep("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestResolveShellUsesBashWhenDeclared(t *testing.T) {
	exec := &Executor{}

	shell := exec.resolveShell(ductfile.Step{Uses: []string{"bash"}})
	assert.NotEmpty(t, shell)
	assert.True(t, strings.Contains(shell, "bash") || shell == "/bin/sh")
}

func TestResolveShellFallbacksFromFish(t *testing.T) {
	exec := &Executor{}
	t.Setenv("SHELL", "/usr/bin/fish")

	shell := exec.resolveShell(ductfile.Step{})
	assert.Equal(t, "/bin/sh", shell)
}
