package migration

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/HikaruEgashira/gh-migrate/acp"
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

	// exec Claude Code with ACP
	promptOption := cmd.Flag("prompt").Value.String()
	autoApprove, _ := cmd.Flags().GetBool("auto-approve")
	if promptOption != "" {
		titleTemplate = titleTemplate + " claude: " + promptOption
		bodyTemplate = bodyTemplate + "\n### Claude Code Prompt\n" + promptOption

		ctx := context.Background()
		if err := acp.RunClaudeSession(ctx, workPath, promptOption, autoApprove); err != nil {
			return fmt.Errorf("Claude Code execution failed: %w", err)
		}
		log.Printf("INFO: Claude Code session completed")
	}

	// detect changed files
	changedFiles, err := getChangedFiles(workPath)
	if err != nil {
		return fmt.Errorf("failed to detect changed files: %w", err)
	}

	if len(changedFiles) == 0 {
		log.Printf("WARN: No changes detected, skipping PR creation")
		return nil
	}

	log.Printf("INFO: Detected %d changed files", len(changedFiles))

	// create branch locally
	if err := runGitCommand(workPath, "checkout", "-b", branchNameTemplate); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	// stage all changes
	if err := runGitCommand(workPath, "add", "-A"); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// commit with signing (uses user's git config)
	if err := runGitCommand(workPath, "commit", "-m", titleTemplate); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	log.Printf("INFO: Created signed commit on branch %s", branchNameTemplate)

	// push to remote
	if err := runGitCommand(workPath, "push", "-u", "origin", branchNameTemplate); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}
	log.Printf("INFO: Pushed branch %s to remote", branchNameTemplate)

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

func runGitCommand(workPath string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = workPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(output))
	}
	return nil
}

func getChangedFiles(workPath string) ([]string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = workPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git status failed: %w", err)
	}

	var files []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		status := line[:2]
		file := strings.TrimSpace(line[3:])
		if file == "" {
			continue
		}
		// handle renamed files (R status shows "old -> new")
		if strings.Contains(file, " -> ") {
			parts := strings.Split(file, " -> ")
			file = parts[len(parts)-1]
		}
		// include modified, added, renamed files (exclude deleted)
		if status[0] != 'D' && status[1] != 'D' {
			files = append(files, file)
		}
	}

	return files, nil
}
