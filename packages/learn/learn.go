package learn

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/HikaruEgashira/gh-migrate/packages/acp"
	"github.com/HikaruEgashira/gh-migrate/packages/tui"
)

type Options struct {
	URL       string
	Name      string
	OutputDir string
}

const learnPromptTemplate = `You are a prompt engineer. Analyze this PR/commit diff and create a reusable Claude Code slash command.

OUTPUT FORMAT (must follow exactly):
---
description: One line description
---
Your prompt content here

RULES:
1. Start with "---" on the first line
2. Include "description: " line
3. End frontmatter with "---" on its own line
4. Write the reusable prompt after the frontmatter
5. Focus on the pattern, not specific files
6. Make it work for any repository

DIFF:
%s

OUTPUT:`

func Execute(opts *Options, ui *tui.UI) error {
	ctx := context.Background()

	ui.Step("parse URL")
	parsed, err := ParseURL(opts.URL)
	if err != nil {
		ui.StepError()
		return err
	}
	ui.StepDone()

	ui.Step("fetch diff")
	diff, err := FetchDiff(parsed)
	if err != nil {
		ui.StepError()
		return err
	}
	ui.StepDone()
	ui.Log("found %d file(s)", len(diff.Files))

	ui.Step("generate prompt (claude code)")
	prompt := fmt.Sprintf(learnPromptTemplate, diff.FormatForPrompt())

	tempDir, err := os.MkdirTemp("", "gh-migrate-learn-*")
	if err != nil {
		ui.StepError()
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	result, err := acp.RunClaudeSession(ctx, tempDir, prompt, true, ui, "")
	if err != nil {
		ui.StepError()
		return fmt.Errorf("Claude Code execution failed: %w", err)
	}
	ui.StepDone()

	ui.Step("save command")
	generated := result.AgentResponse

	filename := opts.Name
	if filename == "" {
		filename = generateFilename(diff.Title)
	}
	if !strings.HasSuffix(filename, ".md") {
		filename += ".md"
	}

	outputPath := filepath.Join(opts.OutputDir, filename)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		ui.StepError()
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(generated), 0o644); err != nil {
		ui.StepError()
		return fmt.Errorf("failed to save command: %w", err)
	}
	ui.StepDone()

	ui.Success("saved: %s", outputPath)
	return nil
}

func generateFilename(title string) string {
	title = strings.ToLower(title)
	re := regexp.MustCompile(`[^a-z0-9\s-]`)
	title = re.ReplaceAllString(title, "")
	title = strings.ReplaceAll(title, " ", "-")
	re = regexp.MustCompile(`-+`)
	title = re.ReplaceAllString(title, "-")
	title = strings.Trim(title, "-")

	if title == "" {
		title = "learned-prompt"
	}

	if len(title) > 50 {
		title = title[:50]
	}

	return title
}
