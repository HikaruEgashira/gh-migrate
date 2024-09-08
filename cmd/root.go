/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	gh "github.com/cli/go-gh/v2"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gh-migrate",
	Short: "PRを作成します",
	Long:  `PRを作成します`,
	Run: func(cmd *cobra.Command, args []string) {
		repo := cmd.Flag("repo").Value.String()
		workPath := os.Getenv("HOME") + "/workspaces/" + repo
		currentPath, _ := os.Getwd()

		titleTemplate := "[gh-migrate]"
		bodyTemplate := `This PR is created by [gh-migrate](https://github.com/HikaruEgashira/gh-migrate)
---`
		branchNameTemplate := "gh-migrate-" + time.Now().Format("20060102150405")

		force := cmd.Flag("force").Value.String()

		if force == "true" {
			err := os.RemoveAll(workPath)
			if err != nil {
				log.Fatal(err)
			}
		}

		_, err := os.Stat(workPath)
		if err != nil {
			cloneArgs := []string{"repo", "clone", repo, workPath, "--", "--depth=1"}
			_, _, err = gh.Exec(cloneArgs...)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Repository cloned: %s\n", repo)
		}
		os.Chdir(workPath)

		// get default branch
		stdout, _, _ := gh.Exec("repo", "view", "--json", "defaultBranchRef", "-q", ".defaultBranchRef.name")
		defaultBranch := strings.TrimSpace(stdout.String())

		// exec command
		cmdOption := cmd.Flag("cmd").Value.String()
		if cmdOption != "" {
			execCommand(cmdOption, &titleTemplate, &bodyTemplate)
		}
		shOption := cmd.Flag("sh").Value.String()
		if shOption != "" {
			execShellScript(shOption, &titleTemplate, &bodyTemplate, currentPath)
		}
		astgrepOption := cmd.Flag("astgrep").Value.String()
		if astgrepOption != "" {
			execAstGrep(astgrepOption, &titleTemplate, &bodyTemplate)
		}
		semgrepOption := cmd.Flag("semgrep").Value.String()
		if semgrepOption != "" {
			execSemgrep(semgrepOption, &titleTemplate, &bodyTemplate)
		}

		// create branch
		exec.Command("git", "switch", "-c", branchNameTemplate).Run()
		exec.Command("git", "add", ".").Run()
		statusOutput, _ := exec.Command("git", "status", "--porcelain").Output()
		fmt.Println(string(statusOutput))
		if len(statusOutput) == 0 {
			fmt.Println("No changes to commit. Exiting.")
			return
		}

		exec.Command("git", "commit", "-m", titleTemplate).Run()
		exec.Command("git", "push", "origin", branchNameTemplate).Run()

		// create PR
		prArgs := []string{
			"pr",
			"create",
			"--base", defaultBranch,
			"--head", branchNameTemplate,
			"--title", titleTemplate,
			"--body", bodyTemplate,
			"--repo", repo,
		}
		stdout, stderr, err := gh.Exec(prArgs...)
		if err != nil {
			fmt.Println(stderr.String())
			log.Fatal(err)
		}
		fmt.Println(stdout.String())
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func execCommand(cmdOption string, titleTemplate *string, bodyTemplate *string) {
	*titleTemplate = *titleTemplate + " run " + cmdOption
	*bodyTemplate = *bodyTemplate + "\n" + cmdOption

	runOutput, err := exec.Command("sh", "-c", cmdOption).CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(runOutput))
}

func execShellScript(shOption string, titleTemplate *string, bodyTemplate *string, currentPath string) {
	scriptFile := "__migrate.sh"
	scriptContent, err := os.ReadFile(currentPath + "/" + shOption)
	if err != nil {
		log.Fatal(err)
	}

	*titleTemplate = *titleTemplate + " run " + shOption
	*bodyTemplate = *bodyTemplate + "\n" + "```bash\n" + string(scriptContent) + "\n```"

	os.WriteFile(scriptFile, scriptContent, 0755)
	runOutput, err := exec.Command("sh", scriptFile).CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(runOutput))

	exec.Command("rm", scriptFile).Run()
}

func execAstGrep(astgrepOption string, titleTemplate *string, bodyTemplate *string) {
	*titleTemplate = *titleTemplate + " run astgrep " + astgrepOption
	*bodyTemplate = *bodyTemplate + "\n" + "```yaml\n" + astgrepOption + "\n```"

	runOutput, err := exec.Command("astgrep", "-c", astgrepOption).CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(runOutput))
}

func execSemgrep(semgrepOption string, titleTemplate *string, bodyTemplate *string) {
	*titleTemplate = *titleTemplate + " run semgrep " + semgrepOption
	*bodyTemplate = *bodyTemplate + "\n" + "```yaml\n" + semgrepOption + "\n```"

	runOutput, err := exec.Command("semgrep", "--config", semgrepOption).CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(runOutput))
}

func init() {
	rootCmd.Flags().StringP("repo", "r", "", "リポジトリ名")
	rootCmd.MarkFlagRequired("repo")
	rootCmd.Flags().BoolP("force", "f", false, "cacheを削除して再取得します")
	rootCmd.Flags().StringP("cmd", "c", "", "引数にあるコマンドを実行します")
	rootCmd.Flags().StringP("sh", "s", "", "引数にあるシェルスクリプトファイルを実行します")
	rootCmd.Flags().StringP("astgrep", "a", "", "引数にあるymlをast-grepとして実行します")
	rootCmd.Flags().StringP("Semgrep", "g", "", "引数にあるymlをsemgrepとして実行します")
}
