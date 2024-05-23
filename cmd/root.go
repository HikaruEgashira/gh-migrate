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
			// remove directory
			err := os.RemoveAll(workPath)
			if err != nil {
				log.Fatal(err)
			}
		}

		// check if directory exists
		_, err := os.Stat(workPath)
		if err != nil {
			cloneArgs := []string{"repo", "clone", repo, workPath, "--", "--depth=1"}
			_, _, err = gh.Exec(cloneArgs...)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Repository cloned: %s\n", repo)
		}

		// get default branch
		os.Chdir(workPath)
		stdout, _, _ := gh.Exec("repo", "view", "--json", "defaultBranchRef", "-q", ".defaultBranchRef.name")
		defaultBranch := strings.TrimSpace(stdout.String())

		// exec command
		cmdOption := cmd.Flag("cmd").Value.String()
		if cmdOption != "" {
			titleTemplate = titleTemplate + " run " + cmdOption
			bodyTemplate = bodyTemplate + "\n" + cmdOption

			runOutput, err := exec.Command("sh", "-c", cmdOption).CombinedOutput()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(runOutput))
		}

		shOption := cmd.Flag("sh").Value.String()
		if shOption != "" {
			scriptFile := "__migrate.sh"
			scriptContent, err := os.ReadFile(currentPath + "/" + shOption)
			if err != nil {
				log.Fatal(err)
			}

			titleTemplate = titleTemplate + " run " + shOption
			bodyTemplate = bodyTemplate + "\n" + "```bash\n" + string(scriptContent) + "\n```"

			os.WriteFile(scriptFile, scriptContent, 0755)
			runOutput, err := exec.Command("sh", scriptFile).CombinedOutput()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(runOutput))

			exec.Command("rm", scriptFile).Run()
		}

		// create branch
		var out []byte
		out, _ = exec.Command("git", "switch", "-c", branchNameTemplate).CombinedOutput()
		fmt.Println(string(out))
		out, _ = exec.Command("git", "add", ".").CombinedOutput()
		fmt.Println(string(out))
		statusOutput, _ := exec.Command("git", "status", "--porcelain").CombinedOutput()
		if len(statusOutput) == 0 {
			fmt.Println("No changes to commit. Exiting.")
			return
		}

		out, _ = exec.Command("git", "commit", "-m", titleTemplate).CombinedOutput()
		fmt.Println(string(out))
		out, _ = exec.Command("git", "push", "origin", branchNameTemplate).CombinedOutput()
		fmt.Println(string(out))

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
		fmt.Println(prArgs)
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

func init() {
	rootCmd.Flags().StringP("repo", "r", "", "リポジトリ名")
	rootCmd.MarkFlagRequired("repo")
	rootCmd.Flags().BoolP("force", "f", false, "cacheを削除して再取得します")
	rootCmd.Flags().StringP("cmd", "c", "", "引数にあるコマンドを実行します")
	rootCmd.Flags().StringP("sh", "s", "", "引数にあるシェルスクリプトファイルを実行します")
}
