package learn

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/HikaruEgashira/gh-migrate/packages/acp"
	"github.com/HikaruEgashira/gh-migrate/packages/tui"
)

type Options struct {
	URL       string
	Name      string
	OutputDir string
}

func buildPrompt(diff *DiffResult, outputDir string, customName string) string {
	filenameInstruction := ""
	if customName != "" {
		filenameInstruction = fmt.Sprintf("Use exactly this filename: %s.md", customName)
	} else {
		filenameInstruction = "Choose a descriptive kebab-case filename based on what the command does (e.g., add-security-policy.md, setup-ci-workflow.md)"
	}

	return fmt.Sprintf(`Analyze this PR/commit diff and create a reusable Claude Code slash command file.

TASK:
1. Analyze the diff to understand what changes were made
2. Create a reusable prompt that can apply similar changes to other repositories
3. Save the file to: %s/<filename>.md

FILE FORMAT (Claude Code slash command):
---
description: One line description of what this command does
---
<Your reusable prompt here>

FILENAME: %s

RULES FOR THE PROMPT:
- Focus on the pattern/intent, not specific file paths
- Make it generalizable to any repository
- Be clear about what should be created or modified
- Include any conventions or best practices observed

DIFF TO ANALYZE:
%s

Create the file now.`, outputDir, filenameInstruction, diff.FormatForPrompt())
}

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

	// Get absolute path for output directory
	outputDir, err := filepath.Abs(opts.OutputDir)
	if err != nil {
		return fmt.Errorf("failed to resolve output dir: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	ui.Step("generate and save command (claude code)")
	prompt := buildPrompt(diff, outputDir, opts.Name)

	_, err = acp.RunClaudeSession(ctx, outputDir, prompt, true, ui, "")
	if err != nil {
		ui.StepError()
		return fmt.Errorf("Claude Code execution failed: %w", err)
	}
	ui.StepDone()

	ui.Success("command saved to: %s", outputDir)
	return nil
}
