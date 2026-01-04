package cli

import (
	"github.com/HikaruEgashira/gh-migrate/packages/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gh-migrate",
	Short: "Creates PRs for specified repositories",
	Long: `gh-migrate is a tool that creates PRs for specified repositories.

Available Commands:
  prompt    Execute Claude Code with a prompt and create a PR
  exec      Execute a command or script and create a PR
  learn     Learn from a PR or commit and generate a reusable prompt

Examples:
  gh migrate prompt "Add gitignore" --repo owner/repo
  gh migrate exec "sed -i 's/old/new/g' file.txt" --repo owner/repo
  gh migrate learn https://github.com/owner/repo/pull/123

For detailed usage examples and flag descriptions, please refer to the README.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		ui := tui.Get()
		if ui != nil {
			ui.Error("%v", err)
			ui.Done()
		}
	}
}

func init() {
}
