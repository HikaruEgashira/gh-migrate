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
	"sync"
	"time"

	gh "github.com/cli/go-gh/v2"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gh-migrate",
	Short: "Creates a PR",
	Long:  `Creates a PR`,
	Run: func(cmd *cobra.Command, args []string) {
		repos := strings.Split(cmd.Flag("repo").Value.String(), ",")
		var wg sync.WaitGroup

		for _, repo := range repos {
			wg.Add(1)
			go func(repo string) {
				defer wg.Done()
				processRepo(repo, cmd)
			}(repo)
		}

		wg.Wait()
	},
}

func processRepo(repo string, cmd *cobra.Command) {
	workPath := os.Getenv("HOME") + "/.gh-migrate/" + repo
	if cmd.Flag("workpath").Value.String() != "" {
		workPath = cmd.Flag("workpath").Value.String() + "/" + repo
	}
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
		execAstGrep(astgrepOption, &titleTemplate, &bodyTemplate, currentPath)
	}
	semgrepOption := cmd.Flag("semgrep").Value.String()
	if semgrepOption != "" {
		execSemgrep(semgrepOption, &titleTemplate, &bodyTemplate, currentPath)
	}

	// create branch
	err = exec.Command("git", "switch", "--create", branchNameTemplate).Run()
	if err != nil {
		log.Fatalf("Failed to create branch: %v", err)
	}
	err = exec.Command("git", "add", ".").Run()
	if err != nil {
		log.Fatalf("Failed to add changes: %v", err)
	}
	statusOutput, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		log.Fatalf("Failed to get git status: %v", err)
	}
	fmt.Println("Git status output:", string(statusOutput))
	if len(statusOutput) == 0 {
		fmt.Println("No changes to commit. Exiting.")
		return
	}

	commitArgs := []string{"commit", "-m", titleTemplate}
	err = exec.Command("git", commitArgs...).Run()
	if err != nil {
		log.Fatalf("Failed to commit changes: %v", err)
	}
	err = exec.Command("git", "push", "-u", "origin", branchNameTemplate).Run()
	if err != nil {
		log.Fatalf("Failed to push changes: %v", err)
	}

	// set static title if flag exists
	if cmd.Flag("title").Value.String() != "" {
		titleTemplate = cmd.Flag("title").Value.String()
	}

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
		fmt.Println("PR creation error:", stderr.String())
		log.Fatal(err)
	}
	fmt.Println(stdout.String())

	// open PR
	if cmd.Flag("open").Value.String() != "" {
		exec.Command("open", stdout.String()).Run()
	}
	if cmd.Flag("with-dev").Value.String() != "" {
		exec.Command("open", strings.ReplaceAll(stdout.String(), "com/", "dev/")).Run()
	}
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

func execAstGrep(astgrepOption string, titleTemplate *string, bodyTemplate *string, currentPath string) {
	scriptContent, err := os.ReadFile(currentPath + "/" + astgrepOption)
	if err != nil {
		log.Fatal(err)
	}

	*titleTemplate = *titleTemplate + " run astgrep " + astgrepOption
	*bodyTemplate = *bodyTemplate + "\n" + "```yaml\n" + string(scriptContent) + "\n```"

	runOutput, err := exec.Command("sg", "scan", "-r", currentPath+"/"+astgrepOption, "--no-ignore", "hidden", "-U").CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(runOutput))
}

func execSemgrep(semgrepOption string, titleTemplate *string, bodyTemplate *string, currentPath string) {
	scriptContent, err := os.ReadFile(currentPath + "/" + semgrepOption)
	if err != nil {
		log.Fatal(err)
	}

	*titleTemplate = *titleTemplate + " run semgrep " + semgrepOption
	*bodyTemplate = *bodyTemplate + "\n" + "```yaml\n" + string(scriptContent) + "\n```"

	runOutput, err := exec.Command("semgrep", "--config", currentPath+"/"+semgrepOption).CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(runOutput))
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
}
