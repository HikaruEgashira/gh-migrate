package cli

import (
	"fmt"
	"strings"
	"sync"

	"github.com/HikaruEgashira/gh-migrate/packages/migration"
	"github.com/HikaruEgashira/gh-migrate/packages/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gh-migrate",
	Short: "Creates a PR",
	Long: `gh-migrate is a tool that creates PRs for specified repositories.
It can be used in the following scenarios:

1. Execute a command and create a PR:
   gh migrate --repo HikaruEgashira/gh-migrate --cmd "sed -i '' 's/gh-migrate/gh-migrate2/g' README.md"

2. Execute a shell script and create a PR:
   gh migrate --repo HikaruEgashira/gh-migrate --sh scripts/update-version.sh

3. Use ast-grep and create a PR:
   gh migrate --repo HikaruEgashira/gh-migrate --astgrep rules/upgrade-actions-checkout.yml

4. Use semgrep and create a PR:
   gh migrate --repo HikaruEgashira/gh-migrate --semgrep rules/security-check.yml

5. Use Claude Code and create a PR:
   gh migrate --repo HikaruEgashira/gh-migrate --prompt "Translate README.md to English"

For detailed usage examples and flag descriptions, please refer to the README.`,
	Run: func(cmd *cobra.Command, args []string) {
		repos := strings.Split(cmd.Flag("repo").Value.String(), ",")

		ui := tui.Init("gh-migrate")
		defer ui.Done()

		ui.Status(fmt.Sprintf("processing %d repo(s)", len(repos)))

		var wg sync.WaitGroup
		errChan := make(chan error, len(repos))

		for _, repo := range repos {
			wg.Add(1)
			go func(repo string) {
				defer wg.Done()
				if err := migration.ExecuteMigration(repo, cmd, ui); err != nil {
					errChan <- fmt.Errorf("%s: %w", repo, err)
				}
			}(repo)
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			ui.Error("%v", err)
		}
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
	rootCmd.Flags().StringP("repo", "r", "", "Specify repository name (multiple repositories can be specified with comma separation)")
	rootCmd.MarkFlagRequired("repo")
	rootCmd.Flags().BoolP("force", "f", false, "Delete cache and re-fetch")
	rootCmd.Flags().StringP("cmd", "c", "", "Execute the command provided as an argument")
	rootCmd.Flags().StringP("sh", "s", "", "Execute the shell script file provided as an argument")
	rootCmd.Flags().StringP("astgrep", "a", "", "Execute the yml file provided as an argument as ast-grep")
	rootCmd.Flags().StringP("semgrep", "g", "", "Execute the yml file provided as an argument as semgrep")
	rootCmd.Flags().String("open", "", "Open the created PR in the browser")
	rootCmd.Flags().String("with-dev", "", "Open the created PR in github.dev")
	rootCmd.Flags().StringP("workpath", "w", "", "Specify the path of the working directory")
	rootCmd.Flags().StringP("title", "t", "", "Specify the title of the PR")
	rootCmd.Flags().StringP("prompt", "P", "", "Execute Claude Code with the prompt provided as an argument")
	rootCmd.Flags().Bool("auto-approve", false, "Auto-approve permission requests from Claude Code")
}
