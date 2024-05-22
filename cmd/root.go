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

		// TODO
		fmt.Println("TODO")
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
}
