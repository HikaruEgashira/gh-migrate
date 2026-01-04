package cli

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/HikaruEgashira/gh-migrate/packages/migration"
	"github.com/HikaruEgashira/gh-migrate/packages/tui"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec <command>",
	Short: "Execute a command or script and create a PR",
	Long: `Execute a shell command or script file and create a PR.

If the argument is a file path, it will be executed as a shell script.
Otherwise, it will be executed as a shell command.

Examples:
  gh migrate exec "sed -i '' 's/old/new/g' README.md" --repo owner/repo
  gh migrate exec scripts/update-version.sh --repo owner/repo --title "Update version"`,
	Args: cobra.ExactArgs(1),
	RunE: runExec,
}

func init() {
	rootCmd.AddCommand(execCmd)

	execCmd.Flags().StringP("repo", "r", "", "Specify repository name (multiple repositories can be specified with comma separation)")
	execCmd.MarkFlagRequired("repo")
	execCmd.Flags().BoolP("force", "f", false, "Delete cache and re-fetch")
	execCmd.Flags().String("open", "", "Open the created PR in the browser")
	execCmd.Flags().String("with-dev", "", "Open the created PR in github.dev")
	execCmd.Flags().StringP("workpath", "w", "", "Specify the path of the working directory")
	execCmd.Flags().StringP("title", "t", "", "Specify the title of the PR (overrides auto-generated title)")
}

func runExec(cmd *cobra.Command, args []string) error {
	command := args[0]

	// オプションの構築
	repo, _ := cmd.Flags().GetString("repo")
	repos := strings.Split(repo, ",")

	workPath, _ := cmd.Flags().GetString("workpath")
	force, _ := cmd.Flags().GetBool("force")
	open, _ := cmd.Flags().GetString("open")
	withDev, _ := cmd.Flags().GetString("with-dev")
	title, _ := cmd.Flags().GetString("title")

	currentPath, _ := os.Getwd()

	ui := tui.Init("gh-migrate exec")
	defer ui.Done()

	ui.Status(fmt.Sprintf("processing %d repo(s)", len(repos)))

	var wg sync.WaitGroup
	errChan := make(chan error, len(repos))

	for _, r := range repos {
		wg.Add(1)
		go func(r string) {
			defer wg.Done()

			opts := migration.MigrationOptions{
				Repo:        r,
				WorkPath:    workPath,
				Force:       force,
				Open:        open != "",
				WithDev:     withDev != "",
				Mode:        migration.ModeExec,
				Command:     command,
				Title:       title,
				CurrentPath: currentPath,
			}

			if err := migration.ExecuteMigration(opts, ui); err != nil {
				errChan <- fmt.Errorf("%s: %w", r, err)
			}
		}(r)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		ui.Error("%v", err)
	}

	return nil
}
