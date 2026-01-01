package acp

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/HikaruEgashira/gh-migrate/tui"
	"github.com/coder/acp-go-sdk"
)

// ClaudeResult contains the result from a Claude Code session
type ClaudeResult struct {
	AgentResponse string
}

func RunClaudeSession(ctx context.Context, workDir string, prompt string, autoApprove bool, ui *tui.UI) (*ClaudeResult, error) {
	cmd := exec.CommandContext(ctx, "npx", "-y", "@zed-industries/claude-code-acp@latest")
	cmd.Dir = workDir
	cmd.Stderr = io.Discard

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe error: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe error: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Claude Code: %w", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
	}()

	client := &MigrationClient{
		AutoApprove: autoApprove,
		WorkDir:     workDir,
		TUI:         ui.GetModel(),
		Program:     ui.GetProgram(),
	}
	conn := acp.NewClientSideConnection(client, stdin, stdout)

	ui.Log("connecting to Claude Code...")

	initResp, err := conn.Initialize(ctx, acp.InitializeRequest{
		ProtocolVersion: acp.ProtocolVersionNumber,
		ClientCapabilities: acp.ClientCapabilities{
			Fs: acp.FileSystemCapability{
				ReadTextFile:  true,
				WriteTextFile: true,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("initialize error: %w", err)
	}
	ui.Log("connected (v%v)", initResp.ProtocolVersion)

	newSess, err := conn.NewSession(ctx, acp.NewSessionRequest{
		Cwd:        workDir,
		McpServers: []acp.McpServer{},
	})
	if err != nil {
		return nil, fmt.Errorf("newSession error: %w", err)
	}
	ui.Log("session: %s", newSess.SessionId[:8])

	// Wrap user prompt with PR body generation instruction
	wrappedPrompt := prompt + `

After completing the task above, please provide a brief summary of what you did for the PR description. Start the summary with "## Summary" and keep it concise (2-3 sentences).`

	_, err = conn.Prompt(ctx, acp.PromptRequest{
		SessionId: newSess.SessionId,
		Prompt:    []acp.ContentBlock{acp.TextBlock(wrappedPrompt)},
	})
	if err != nil {
		return nil, fmt.Errorf("prompt error: %w", err)
	}

	return &ClaudeResult{
		AgentResponse: client.GetAgentResponse(),
	}, nil
}
