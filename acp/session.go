package acp

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/HikaruEgashira/gh-migrate/tui"
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

	// Check if TTY is available
	isTTY := term.IsTerminal(int(os.Stderr.Fd()))

	var tuiModel *tui.Model
	var program *tea.Program

	if isTTY {
		tuiModel = tui.New("Claude Code")
		program = tea.NewProgram(tuiModel, tea.WithOutput(os.Stderr))
	}

	client := &MigrationClient{
		AutoApprove: autoApprove,
		WorkDir:     workDir,
		TUI:         tuiModel,
		Program:     program,
	}
	conn := acp.NewClientSideConnection(client, stdin, stdout)

	if isTTY {
		// Run TUI in background
		go func() {
			if _, err := program.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			}
		}()
		tuiModel.SendUpdate(program, "status", "", "connecting", "")
	}

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
		if tuiModel != nil {
			tuiModel.Complete(program)
		}
		return fmt.Errorf("initialize error: %w", err)
	}
	if tuiModel != nil {
		tuiModel.SendUpdate(program, "status", "", fmt.Sprintf("connected (v%v)", initResp.ProtocolVersion), "")
	}

	newSess, err := conn.NewSession(ctx, acp.NewSessionRequest{
		Cwd:        workDir,
		McpServers: []acp.McpServer{},
	})
	if err != nil {
		if tuiModel != nil {
			tuiModel.Complete(program)
		}
		return fmt.Errorf("newSession error: %w", err)
	}
	if tuiModel != nil {
		tuiModel.SendUpdate(program, "status", "", "running", "")
		tuiModel.SendUpdate(program, "output", "", "", fmt.Sprintf("session: %s", newSess.SessionId[:8]))
	}

	_, err = conn.Prompt(ctx, acp.PromptRequest{
		SessionId: newSess.SessionId,
		Prompt:    []acp.ContentBlock{acp.TextBlock(prompt)},
	})
	if err != nil {
		if tuiModel != nil {
			tuiModel.Complete(program)
		}
		return fmt.Errorf("prompt error: %w", err)
	}

	if tuiModel != nil {
		tuiModel.Complete(program)
	}
	return nil
}
