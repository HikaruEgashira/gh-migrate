package migration

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/HikaruEgashira/gh-migrate/scripts"
	gh "github.com/cli/go-gh/v2"
	"github.com/spf13/cobra"
	"github.com/cli/go-gh/v2/pkg/api")

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

	// Initialize: fetch latest and reset to default branch
	stdout, _, _ := gh.Exec("repo", "view", "--json", "defaultBranchRef", "-q", ".defaultBranchRef.name")
	defaultBranch := strings.TrimSpace(stdout.String())

	// Fetch latest from remote
	fetchCmd := exec.Command("git", "fetch", "origin", defaultBranch)
	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch latest: %v", err)
	}

	// Checkout and reset to latest default branch
	checkoutCmd := exec.Command("git", "checkout", defaultBranch)
	if err := checkoutCmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout default branch: %v", err)
	}

	resetCmd := exec.Command("git", "reset", "--hard", "origin/"+defaultBranch)
	if err := resetCmd.Run(); err != nil {
		return fmt.Errorf("failed to reset to latest: %v", err)
	}
	log.Printf("INFO: Reset to latest %s", defaultBranch)

	// get default branch SHA
	stdout, _, _ = gh.Exec("api", fmt.Sprintf("repos/%s/git/refs/heads/%s", repo, defaultBranch))
	var refResponse struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	if err := json.Unmarshal([]byte(stdout.String()), &refResponse); err != nil {
		return fmt.Errorf("failed to parse default branch SHA: %v", err)
	}
	defaultBranchSHA := refResponse.Object.SHA

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

	// create commit using GitHub API
	client, err := api.DefaultRESTClient()
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %v", err)
	}

	// create branch
	path := fmt.Sprintf("repos/%s/git/refs", repo)
	branchPayload := map[string]interface{}{
		"ref": fmt.Sprintf("refs/heads/%s", branchNameTemplate),
		"sha": defaultBranchSHA,
	}
	branchPayloadBytes, err := json.Marshal(branchPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal branch payload: %v", err)
	}
	err = client.Post(path, bytes.NewReader(branchPayloadBytes), nil)
	if err != nil {
		return fmt.Errorf("failed to create branch: %v", err)
	}

	// create commit
	path = fmt.Sprintf("repos/%s/contents/test.txt", repo)
	content := []byte("test\n")
	encodedContent := make([]byte, base64.StdEncoding.EncodedLen(len(content)))
	base64.StdEncoding.Encode(encodedContent, content)

	commitPayload := map[string]interface{}{
		"message": titleTemplate,
		"branch":  branchNameTemplate,
		"content": string(encodedContent),
	}
	commitPayloadBytes, err := json.Marshal(commitPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal commit payload: %v", err)
	}
	err = client.Put(path, bytes.NewReader(commitPayloadBytes), nil)
	if err != nil {
		return fmt.Errorf("failed to create commit: %v", err)
	}
	log.Printf("INFO: Created commit on branch %s", branchNameTemplate)

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
