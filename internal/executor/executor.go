package executor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/nmvinicius/duct/internal/config"
	"github.com/nmvinicius/duct/internal/parser"
	"github.com/nmvinicius/duct/internal/platform"
	"github.com/nmvinicius/duct/pkg/ductfile"
)

// Executor runs pipeline steps
type Executor struct {
	df       *ductfile.Ductfile
	platform platform.Platform
	config   *config.Config
	env      map[string]string
}

// New creates a new Executor
func New(df *ductfile.Ductfile, plat platform.Platform, cfg *config.Config) *Executor {
	env := make(map[string]string)

	// Add globals
	for k, v := range df.Globals {
		env[k] = v
	}

	// Add git vars
	for k, v := range plat.GitVars() {
		env[k] = v
	}

	// Add project
	env["PROJECT"] = df.Project
	env["DUCT_VERSION"] = "0.1.0"

	if plat == platform.Local {
		if env["GIT_COMMIT"] == "" {
			if out, err := exec.Command("git", "rev-parse", "HEAD").Output(); err == nil {
				env["GIT_COMMIT"] = strings.TrimSpace(string(out))
			}
		}
		if env["GIT_BRANCH"] == "" {
			if out, err := exec.Command("git", "branch", "--show-current").Output(); err == nil {
				env["GIT_BRANCH"] = strings.TrimSpace(string(out))
			}
		}
		if env["GIT_TAG"] == "" {
			if out, err := exec.Command("git", "describe", "--tags", "--exact-match").Output(); err == nil {
				env["GIT_TAG"] = strings.TrimSpace(string(out))
			}
		}
	}

	return &Executor{
		df:       df,
		platform: plat,
		config:   cfg,
		env:      env,
	}
}

// RunAll executes all steps in dependency order
func (e *Executor) RunAll() error {
	steps, err := parser.GetExecutionOrder(e.df.Steps)
	if err != nil {
		return err
	}

	color.Cyan("Executing %d steps...", len(steps))

	for _, step := range steps {
		if err := e.runStep(step); err != nil {
			if !step.AllowFail {
				// Try rollback
				if step.Rollback != "" {
					color.Yellow("Rolling back step %s...", step.Name)
					e.runCommand(step.Rollback, step.Name+"-rollback", e.resolveShell(step))
				}
				return fmt.Errorf("step %q failed: %w", step.Name, err)
			}
			color.Yellow("Step %s failed but ALLOW_FAIL is set, continuing...", step.Name)
		}
	}

	color.Green("✅ Pipeline completed successfully!")
	return nil
}

// RunStep executes a single step by name
func (e *Executor) RunStep(name string) error {
	for _, s := range e.df.Steps {
		if s.Name == name {
			return e.runStep(s)
		}
	}
	return fmt.Errorf("step %q not found", name)
}

func (e *Executor) runStep(step ductfile.Step) error {
	color.Cyan("\n▶ Step: %s", step.Name)

	// Check condition
	if step.When != nil {
		branch := e.env["GIT_BRANCH"]
		tag := e.env["GIT_TAG"]
		plat := e.platform.String()

		ok, err := step.When.Evaluate(branch, tag, plat)
		if err != nil {
			return fmt.Errorf("condition error: %w", err)
		}
		if !ok {
			color.Yellow("  SKIPPED (condition not met)")
			return nil
		}
	}

	// Setup tools (USE)
	for _, tool := range step.Uses {
		if err := e.setupTool(tool); err != nil {
			return fmt.Errorf("setup tool %q: %w", tool, err)
		}
	}

	shell := e.resolveShell(step)

	// Run commands
	for _, cmd := range step.Runs {
		// Expand variables
		cmd = e.expandVars(cmd)

		color.White("  $ %s", cmd)
		if err := e.runCommand(cmd, step.Name, shell); err != nil {
			return err
		}
	}

	color.Green("  ✓ %s completed", step.Name)
	return nil
}

func (e *Executor) runCommand(cmd, context, shell string) error {
	command := exec.Command(shell, "-c", cmd)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Dir = "."

	// Set environment
	command.Env = os.Environ()
	for k, v := range e.env {
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
	}

	return command.Run()
}

func (e *Executor) resolveShell(step ductfile.Step) string {
	for _, tool := range step.Uses {
		if strings.EqualFold(tool, "bash") {
			if bashPath, err := exec.LookPath("bash"); err == nil {
				return bashPath
			}
			return "/bin/sh"
		}
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		return "/bin/sh"
	}

	if strings.HasSuffix(shell, "fish") || strings.Contains(shell, "fish") {
		return "/bin/sh"
	}

	return shell
}

func (e *Executor) setupTool(tool string) error {
	color.White("  [setup] %s", tool)

	// Try to find setup script
	scriptPath := fmt.Sprintf("%s/tools/%s.sh", e.config.ScriptsDir, tool)
	if _, err := os.Stat(scriptPath); err == nil {
		cmd := exec.Command("/bin/sh", scriptPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		return cmd.Run()
	}

	// Try generic setup
	genericPath := fmt.Sprintf("%s/tools/generic.sh", e.config.ScriptsDir)
	if _, err := os.Stat(genericPath); err == nil {
		cmd := exec.Command("/bin/sh", genericPath, tool)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		return cmd.Run()
	}

	// Fallback: just check if command exists
	if _, err := exec.LookPath(tool); err != nil {
		return fmt.Errorf("tool %q not found and no setup script available", tool)
	}

	return nil
}

func (e *Executor) expandVars(s string) string {
	for k, v := range e.env {
		s = strings.ReplaceAll(s, "$"+k, v)
		s = strings.ReplaceAll(s, "${"+k+"}", v)
	}
	return s
}

// DryRun shows what would execute without running
func DryRun(df *ductfile.Ductfile, plat platform.Platform, cfg *config.Config) error {
	steps, err := parser.GetExecutionOrder(df.Steps)
	if err != nil {
		return err
	}

	env := make(map[string]string)
	for k, v := range df.Globals {
		env[k] = v
	}
	for k, v := range plat.GitVars() {
		env[k] = v
	}
	env["PROJECT"] = df.Project

	fmt.Println()
	for _, step := range steps {
		color.Cyan("▶ %s", step.Name)

		if len(step.Needs) > 0 {
			color.White("  needs: %v", step.Needs)
		}

		if step.When != nil {
			color.Yellow("  when: %s", step.When.Raw)
		}

		for _, tool := range step.Uses {
			color.White("  [use] %s", tool)
		}

		for _, cmd := range step.Runs {
			expanded := cmd
			for k, v := range env {
				expanded = strings.ReplaceAll(expanded, "$"+k, v)
				expanded = strings.ReplaceAll(expanded, "${"+k+"}", v)
			}
			color.White("  $ %s", expanded)
		}

		if step.Rollback != "" {
			color.Yellow("  rollback: %s", step.Rollback)
		}

		fmt.Println()
	}

	return nil
}
