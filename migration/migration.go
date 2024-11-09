package migration

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/HikaruEgashira/gh-migrate/scripts"
	gh "github.com/cli/go-gh/v2"
	"github.com/spf13/cobra"
)

func ExecuteMigration(repo string, cmd *cobra.Command) error {
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
		if err := scripts.ExecCommand(cmdOption, &titleTemplate, &bodyTemplate); err != nil {
			return err
		}
	}
	shOption := cmd.Flag("sh").Value.String()
	if shOption != "" {
		if err := scripts.ExecScript(shOption, &titleTemplate, &bodyTemplate, currentPath, "sh"); err != nil {
			return err
		}
	}
	astgrepOption := cmd.Flag("astgrep").Value.String()
	if astgrepOption != "" {
		if err := scripts.ExecScript(astgrepOption, &titleTemplate, &bodyTemplate, currentPath, "astgrep"); err != nil {
			return err
		}
	}
	semgrepOption := cmd.Flag("semgrep").Value.String()
	if semgrepOption != "" {
		if err := scripts.ExecScript(semgrepOption, &titleTemplate, &bodyTemplate, currentPath, "semgrep"); err != nil {
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
