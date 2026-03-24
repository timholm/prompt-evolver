package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/timholm/prompt-evolver/internal/analyze"
	"github.com/timholm/prompt-evolver/internal/config"
	"github.com/timholm/prompt-evolver/internal/deploy"
	"github.com/timholm/prompt-evolver/internal/evolve"
	"github.com/timholm/prompt-evolver/internal/test"
)

func main() {
	root := &cobra.Command{
		Use:   "prompt-evolver",
		Short: "Autonomous prompt evolution for claude-code-factory",
		Long:  "Analyzes shipped vs failed builds, identifies prompt weaknesses, generates improved prompts, and A/B tests them.",
	}

	root.AddCommand(analyzeCmd())
	root.AddCommand(evolveCmd())
	root.AddCommand(testCmd())
	root.AddCommand(deployCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func analyzeCmd() *cobra.Command {
	var dbPath string
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze build outcomes and extract failure patterns",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			if dbPath != "" {
				cfg.FactoryDataDir = dbPath
			}
			a := analyze.New(cfg)
			report, err := a.Run()
			if err != nil {
				return fmt.Errorf("analyze: %w", err)
			}
			fmt.Println(report.String())
			return nil
		},
	}
	cmd.Flags().StringVar(&dbPath, "db", "", "Override path to factory data directory")
	return cmd
}

func evolveCmd() *cobra.Command {
	var analysisFile string
	cmd := &cobra.Command{
		Use:   "evolve",
		Short: "Generate improved prompts based on failure analysis",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			e := evolve.New(cfg)
			result, err := e.Run(analysisFile)
			if err != nil {
				return fmt.Errorf("evolve: %w", err)
			}
			fmt.Println(result.String())
			return nil
		},
	}
	cmd.Flags().StringVar(&analysisFile, "analysis", "", "Path to analysis JSON (default: run analyze first)")
	return cmd
}

func testCmd() *cobra.Command {
	var count int
	cmd := &cobra.Command{
		Use:   "test",
		Short: "A/B test old vs new prompts on sample builds",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			t := test.New(cfg)
			result, err := t.Run(count)
			if err != nil {
				return fmt.Errorf("test: %w", err)
			}
			fmt.Println(result.String())
			return nil
		},
	}
	cmd.Flags().IntVar(&count, "count", 3, "Number of projects to build per prompt set")
	return cmd
}

func deployCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy improved prompts to claude-code-factory",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			d := deploy.New(cfg)
			err := d.Run(force)
			if err != nil {
				return fmt.Errorf("deploy: %w", err)
			}
			fmt.Println("Prompts deployed successfully.")
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Deploy even if test results are inconclusive")
	return cmd
}
