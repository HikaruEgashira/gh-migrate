/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"
	"os/exec"

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

		// exec command
		cmdOption := cmd.Flag("cmd").Value.String()
		if cmdOption != "" {
			// move target workspace
			err := os.Chdir("workspaces/" + repo)
			if err != nil {
				log.Fatal(err)
			}

			// exec command
			cmd := exec.Command("sh", "-c", cmdOption)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				log.Fatal(err)
			}
		}

		// create branch
		branchArgs := []string{"branch", "feature"}
		_, _, err = gh.Exec(branchArgs...)
		if err != nil {
			log.Fatal(err)
		}

		// push branch
		pushArgs := []string{"push", "origin", "feature"}
		_, _, err = gh.Exec(pushArgs...)
		if err != nil {
			log.Fatal(err)
		}

		// create PR
		prArgs := []string{"pr", "create", "--base", "main", "--head", "feature", "--title", "feature", "--body", "feature"}
		_, _, err = gh.Exec(prArgs...)
		if err != nil {
			log.Fatal(err)
		}

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
}
