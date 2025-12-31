package acp

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/coder/acp-go-sdk"
)

func RunClaudeSession(ctx context.Context, workDir string, prompt string, autoApprove bool) error {
	cmd := exec.CommandContext(ctx, "npx", "-y", "@zed-industries/claude-code-acp@latest")
	cmd.Dir = workDir
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe error: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe error: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Claude Code: %w", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
	}()

	client := &MigrationClient{
		AutoApprove: autoApprove,
		WorkDir:     workDir,
	}
	conn := acp.NewClientSideConnection(client, stdin, stdout)

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
		return fmt.Errorf("initialize error: %w", err)
	}
	log.Printf("[ACP] Connected to Claude Code (protocol v%v)", initResp.ProtocolVersion)

	newSess, err := conn.NewSession(ctx, acp.NewSessionRequest{
		Cwd:        workDir,
		McpServers: []acp.McpServer{},
	})
	if err != nil {
		return fmt.Errorf("newSession error: %w", err)
	}
	log.Printf("[ACP] Created session: %s", newSess.SessionId)

	_, err = conn.Prompt(ctx, acp.PromptRequest{
		SessionId: newSess.SessionId,
		Prompt:    []acp.ContentBlock{acp.TextBlock(prompt)},
	})
	if err != nil {
		return fmt.Errorf("prompt error: %w", err)
	}

	log.Printf("[ACP] Session completed successfully")
	return nil
}
