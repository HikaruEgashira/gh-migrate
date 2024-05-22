/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
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
		force := cmd.Flag("force").Value.String()

		titleTemplate := "[gh-migrate]"
		bodyTemplate := `This PR is created by [gh-migrate](https://github.com/HikaruEgashira/gh-migrate).
		---`
		timestamp := time.Now().Format("20060102150405")
		branchNameTemplate := "gh-migrate-" + timestamp
		defaultBranch := "main"

		if force == "true" {
			// remove directory
			err := os.RemoveAll("workspaces/" + repo)
			if err != nil {
				log.Fatal(err)
			}
		}

		// check if directory exists
		_, err := os.Stat("workspaces/" + repo)
		if err != nil {
			cloneArgs := []string{"repo", "clone", repo, "workspaces/" + repo, "--", "--depth=1"}
			_, _, err = gh.Exec(cloneArgs...)
			if err != nil {
				log.Fatal(err)
			}
		}

		// move target workspace
		err = os.Chdir("workspaces/" + repo)
		if err != nil {
			log.Fatal(err)
		}

		// get default branch
		stdout, _, _ := gh.Exec("repo", "view", "--json", "defaultBranchRef")
		var defaultBranchRef map[string]interface{}
		err = json.Unmarshal(stdout.Bytes(), &defaultBranchRef)
		if err != nil {
			log.Fatal(err)
		}
		defaultBranch = defaultBranchRef["defaultBranchRef"].(map[string]interface{})["name"].(string)

		// exec command
		cmdOption := cmd.Flag("cmd").Value.String()
		if cmdOption != "" {
			titleTemplate = titleTemplate + " " + cmdOption
			bodyTemplate = bodyTemplate + "\n" + cmdOption

			exec.Command("sh", "-c", cmdOption).Run()
		}

		shOption := cmd.Flag("sh").Value.String()
		if shOption != "" {
			titleTemplate = titleTemplate + " " + shOption
			bodyTemplate = bodyTemplate + "\n" + shOption

			scriptContent := []byte(shOption)
			scriptFile := "__migrate.sh"
			err := os.WriteFile(scriptFile, scriptContent, 0755)
			if err != nil {
				log.Fatal(err)
			}
			exec.Command("sh", scriptFile).Run()
			exec.Command("rm", scriptFile).Run()
		}

		exec.Command("git", "switch", "-c", branchNameTemplate).Run()
		exec.Command("git", "add", ".").Run()
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
		fmt.Println(stderr.String())
		if err != nil {
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
