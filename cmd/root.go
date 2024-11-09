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
	Long:  `Creates a PR`,
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
				if err := processRepo(repo, cmd); err != nil {
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

func processRepo(repo string, cmd *cobra.Command) error {
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
			return err
		}
	}

	_, err := os.Stat(workPath)
	if err != nil {
		cloneArgs := []string{"repo", "clone", repo, workPath, "--", "--depth=1"}
		_, _, err = gh.Exec(cloneArgs...)
		if err != nil {
			return err
		}
		log.Printf("INFO: Repository cloned: %s", repo)
	}
	os.Chdir(workPath)

	// get default branch
	stdout, _, _ := gh.Exec("repo", "view", "--json", "defaultBranchRef", "-q", ".defaultBranchRef.name")
	defaultBranch := strings.TrimSpace(stdout.String())

	// exec command
	cmdOption := cmd.Flag("cmd").Value.String()
	if cmdOption != "" {
		if err := execCommand(cmdOption, &titleTemplate, &bodyTemplate); err != nil {
			return err
		}
	}
	shOption := cmd.Flag("sh").Value.String()
	if shOption != "" {
		if err := execScript(shOption, &titleTemplate, &bodyTemplate, currentPath, "sh"); err != nil {
			return err
		}
	}
	astgrepOption := cmd.Flag("astgrep").Value.String()
	if astgrepOption != "" {
		if err := execScript(astgrepOption, &titleTemplate, &bodyTemplate, currentPath, "astgrep"); err != nil {
			return err
		}
	}
	semgrepOption := cmd.Flag("semgrep").Value.String()
	if semgrepOption != "" {
		if err := execScript(semgrepOption, &titleTemplate, &bodyTemplate, currentPath, "semgrep"); err != nil {
			return err
		}
	}

	// create branch
	err = exec.Command("git", "switch", "--create", branchNameTemplate).Run()
	if err != nil {
		return fmt.Errorf("failed to create branch: %v", err)
	}
	err = exec.Command("git", "add", ".").Run()
	if err != nil {
		return fmt.Errorf("failed to add changes: %v", err)
	}
	statusOutput, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		return fmt.Errorf("failed to get git status: %v", err)
	}
	log.Printf("INFO: Git status output: %s", string(statusOutput))
	if len(statusOutput) == 0 {
		log.Println("INFO: No changes to commit. Exiting.")
		return nil
	}

	commitArgs := []string{"commit", "-m", titleTemplate}
	err = exec.Command("git", commitArgs...).Run()
	if err != nil {
		return fmt.Errorf("failed to commit changes: %v", err)
	}
	err = exec.Command("git", "push", "-u", "origin", branchNameTemplate).Run()
	if err != nil {
		return fmt.Errorf("failed to push changes: %v", err)
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
		log.Printf("ERROR: PR creation error: %s", stderr.String())
		return err
	}
	log.Printf("INFO: %s", stdout.String())

	// open PR
	if cmd.Flag("open").Value.String() != "" {
		exec.Command("open", stdout.String()).Run()
	}
	if cmd.Flag("with-dev").Value.String() != "" {
		exec.Command("open", strings.ReplaceAll(stdout.String(), "com/", "dev/")).Run()
	}

	return nil
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}

func execCommand(cmdOption string, titleTemplate *string, bodyTemplate *string) error {
	*titleTemplate = *titleTemplate + " run " + cmdOption
	*bodyTemplate = *bodyTemplate + "\n" + cmdOption

	runOutput, err := exec.Command("sh", "-c", cmdOption).CombinedOutput()
	if err != nil {
		return err
	}
	log.Printf("INFO: %s", string(runOutput))
	return nil
}

func execScript(scriptOption string, titleTemplate *string, bodyTemplate *string, currentPath string, scriptType string) error {
	scriptContent, err := os.ReadFile(currentPath + "/" + scriptOption)
	if err != nil {
		return err
	}

	*titleTemplate = *titleTemplate + " run " + scriptType + " " + scriptOption
	*bodyTemplate = *bodyTemplate + "\n" + "```" + scriptType + "\n" + string(scriptContent) + "\n```"

	var runOutput []byte
	switch scriptType {
	case "sh":
		runOutput, err = exec.Command("sh", currentPath+"/"+scriptOption).CombinedOutput()
	case "astgrep":
		runOutput, err = exec.Command("sg", "scan", "-r", currentPath+"/"+scriptOption, "--no-ignore", "hidden", "-U").CombinedOutput()
	case "semgrep":
		runOutput, err = exec.Command("semgrep", "--config", currentPath+"/"+scriptOption).CombinedOutput()
	}

	if err != nil {
		return err
	}
	log.Printf("INFO: %s", string(runOutput))
	return nil
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
}
