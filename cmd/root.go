/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	gh "github.com/cli/go-gh/v2"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gh-migrate",
	Short: "migrate.md を表示します",
	Long:  `migrate.md を表示します`,
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

		// read migrate.md
		migrate, err := os.ReadFile("workspaces/" + repo + "/migrate.md")
		if err != nil {
			fmt.Println("migrate.md not found")
		}
		fmt.Println(string(migrate))
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
	rootCmd.Flags().BoolP("force", "f", false, "migrate.md を強制的に取得します")
}
