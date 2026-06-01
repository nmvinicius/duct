package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/nmvinicius/duct/internal/config"
	"github.com/nmvinicius/duct/internal/executor"
	"github.com/nmvinicius/duct/internal/parser"
	"github.com/nmvinicius/duct/internal/platform"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "unknown"

	// Flags
	ductfilePath string
	localMode    bool
	dryRun       bool
	debugMode    bool
	stepName     string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "duct",
		Short: "Duct - Pipeline as Code",
		Long: color.CyanString(`
  ██████╗ ██╗   ██╗ ██████╗████████╗
  ██╔══██╗██║   ██║██╔════╝╚══██╔══╝
  ██║  ██║██║   ██║██║        ██║   
  ██║  ██║██║   ██║██║        ██║   
  ██████╔╝╚██████╔╝╚██████╗   ██║   
  ╚═════╝  ╚═════╝  ╚═════╝   ╚═╝   
                                     
`) + "Pipeline as Code — Define, extend and run CI/CD anywhere.",
		Version: fmt.Sprintf("%s (commit: %s)", version, commit),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if debugMode {
				os.Setenv("DUCT_DEBUG", "true")
			}
			return nil
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&ductfilePath, "file", "f", "Ductfile", "Path to Ductfile")
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "Enable debug output")

	// Subcommands
	rootCmd.AddCommand(runCmd())
	rootCmd.AddCommand(validateCmd())
	rootCmd.AddCommand(graphCmd())
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}
}

func runCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [step-name]",
		Short: "Run the pipeline or a specific step",
		Long:  "Execute the Ductfile pipeline. Use --local to simulate CI locally.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				stepName = args[0]
			}

			// Load config
			cfg := config.Load(ductfilePath, localMode)

			// Detect platform
			plat := platform.Detect()
			if localMode {
				plat = platform.Local
				color.Yellow("Running in LOCAL mode")
			}

			color.Cyan("Platform: %s", plat.String())
			color.Cyan("Ductfile: %s", ductfilePath)

			// Parse Ductfile
			df, err := parser.ParseFile(ductfilePath)
			if err != nil {
				return fmt.Errorf("failed to parse Ductfile: %w", err)
			}

			color.Green("Project: %s", df.Project)
			color.Green("Steps: %d", len(df.Steps))

			if dryRun {
				color.Yellow("DRY RUN — no commands will be executed")
				return executor.DryRun(df, plat, cfg)
			}

			// Execute
			exec := executor.New(df, plat, cfg)
			if stepName != "" {
				return exec.RunStep(stepName)
			}
			return exec.RunAll()
		},
	}

	cmd.Flags().BoolVarP(&localMode, "local", "l", false, "Run locally (simulates CI environment)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would run without executing")

	return cmd
}

func validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate Ductfile syntax",
		RunE: func(cmd *cobra.Command, args []string) error {
			color.Cyan("Validating %s...", ductfilePath)

			df, err := parser.ParseFile(ductfilePath)
			if err != nil {
				color.Red("❌ Invalid")
				return err
			}

			// Validate step dependencies
			if _, err := parser.ValidateSteps(df.Steps); err != nil {
				color.Red("❌ Invalid")
				return err
			}

			color.Green("✅ Valid Ductfile")
			color.Green("   Project: %s", df.Project)
			color.Green("   Version: %s", df.Version)
			color.Green("   Steps: %d", len(df.Steps))

			for _, s := range df.Steps {
				status := "  "
				if len(s.Needs) > 0 {
					status = fmt.Sprintf("→ needs: %v", s.Needs)
				}
				color.White("   • %s %s", s.Name, color.CyanString(status))
			}

			return nil
		},
	}
}

func graphCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "graph",
		Short: "Show pipeline dependency graph",
		RunE: func(cmd *cobra.Command, args []string) error {
			df, err := parser.ParseFile(ductfilePath)
			if err != nil {
				return err
			}

			color.Cyan("Dependency Graph for %s", df.Project)
			fmt.Println()

			sorted, err := parser.GetExecutionOrder(df.Steps)
			if err != nil {
				return err
			}

			for i, s := range sorted {
				indent := ""
				for j := 0; j < i; j++ {
					indent += "  "
				}
				arrow := "├─"
				if i == len(sorted)-1 {
					arrow = "└─"
				}

				stepColor := color.New(color.FgWhite)
				if len(s.Needs) == 0 {
					stepColor = color.New(color.FgGreen)
				}

				fmt.Printf("%s%s%s %s\n", indent, arrow, stepColor.Sprint(s.Name), color.CyanString(strings.Join(s.Needs, ", ")))
			}

			return nil
		},
	}
}

func initCmd() *cobra.Command {
	var template string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Ductfile",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat("Ductfile"); err == nil {
				return fmt.Errorf("Ductfile already exists")
			}

			content := `VERSION 1.0

PROJECT my-project

STEP build
    USE node
    RUN echo "Hello from Duct!"
`

			if template != "" {
				// TODO: Load from template
				color.Yellow("Template loading not yet implemented, using default")
			}

			if err := os.WriteFile("Ductfile", []byte(content), 0644); err != nil {
				return err
			}

			color.Green("✅ Created Ductfile")
			color.White("Edit it and run: duct run --local")

			return nil
		},
	}

	cmd.Flags().StringVarP(&template, "template", "t", "", "Template to use (e.g., node, docker)")

	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Duct %s\n", version)
			fmt.Printf("Commit: %s\n", commit)
			fmt.Printf("Go: %s\n", "1.22+")
		},
	}
}
