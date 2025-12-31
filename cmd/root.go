package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/HikaruEgashira/gh-migrate/migration"
	"github.com/spf13/cobra"
)

var (
	logFile *os.File
)

func initLog(output string) {
	var err error
	if output == "stdout" {
		log.SetOutput(os.Stdout)
	} else {
		logFile, err = os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		log.SetOutput(logFile)
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

var rootCmd = &cobra.Command{
	Use:   "gh-migrate",
	Short: "Creates a PR",
	Long: `gh-migrateは、指定されたリポジトリに対してPRを作成するツールです。
以下のようなシナリオで使用できます:

1. コマンドを実行してPRを作成する:
   gh migrate --repo HikaruEgashira/gh-migrate --cmd "sed -i '' 's/gh-migrate/gh-migrate2/g' README.md"

2. シェルスクリプトを実行してPRを作成する:
   gh migrate --repo HikaruEgashira/gh-migrate --sh scripts/update-version.sh

3. ast-grepを使用してPRを作成する:
   gh migrate --repo HikaruEgashira/gh-migrate --astgrep rules/upgrade-actions-checkout.yml

4. semgrepを使用してPRを作成する:
   gh migrate --repo HikaruEgashira/gh-migrate --semgrep rules/security-check.yml

5. Claude Codeを使用してPRを作成する:
   gh migrate --repo HikaruEgashira/gh-migrate --prompt "README.mdを英語に翻訳して"

詳細な使用例やフラグの説明については、READMEをご覧ください。`,
	Run: func(cmd *cobra.Command, args []string) {
		initLog(cmd.Flag("log-output").Value.String())
		defer logFile.Close()

		repos := strings.Split(cmd.Flag("repo").Value.String(), ",")
		var wg sync.WaitGroup
		errChan := make(chan error, len(repos))

		for _, repo := range repos {
			wg.Add(1)
			go func(repo string) {
				defer wg.Done()
				if err := migration.ExecuteMigration(repo, cmd); err != nil {
					errChan <- fmt.Errorf("error processing repo %s: %w", repo, err)
				}
			}(repo)
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			log.Printf("ERROR: %v", err)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("ERROR: %v", err)
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
	rootCmd.Flags().StringP("log-output", "l", "stdout", "Specify the log output (stdout or file path)")
	rootCmd.Flags().StringP("prompt", "P", "", "Execute Claude Code with the prompt provided as an argument")
	rootCmd.Flags().Bool("auto-approve", false, "Auto-approve permission requests from Claude Code")
}
