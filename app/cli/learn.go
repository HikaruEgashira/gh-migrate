package cli

import (
	"github.com/HikaruEgashira/gh-migrate/packages/learn"
	"github.com/HikaruEgashira/gh-migrate/packages/tui"
	"github.com/spf13/cobra"
)

var learnCmd = &cobra.Command{
	Use:   "learn <url>",
	Short: "Learn from a PR or commit and generate a reusable prompt",
	Long: `Learn analyzes a PR or commit diff and generates a Claude Code slash command.

Supported URL formats:
  - PR: https://github.com/owner/repo/pull/123
  - Commit: https://github.com/owner/repo/commit/abc1234

Examples:
  gh migrate learn https://github.com/owner/repo/pull/123
  gh migrate learn https://github.com/owner/repo/commit/abc --name "add-gitignore"`,
	Args: cobra.ExactArgs(1),
	Run:  runLearn,
}

func init() {
	rootCmd.AddCommand(learnCmd)
	learnCmd.Flags().StringP("name", "n", "", "Command name for the slash command file")
	learnCmd.Flags().StringP("output", "o", ".claude/commands", "Output directory for the generated command")
}

func runLearn(cmd *cobra.Command, args []string) {
	ui := tui.Init("gh-migrate learn")
	defer ui.Done()

	name, _ := cmd.Flags().GetString("name")
	output, _ := cmd.Flags().GetString("output")

	opts := &learn.Options{
		URL:       args[0],
		Name:      name,
		OutputDir: output,
	}

	if err := learn.Execute(opts, ui); err != nil {
		ui.Error("%v", err)
	}
}
