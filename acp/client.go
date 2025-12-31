package acp

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/coder/acp-go-sdk"
)

type MigrationClient struct {
	AutoApprove bool
	WorkDir     string
}

var _ acp.Client = (*MigrationClient)(nil)

func (c *MigrationClient) RequestPermission(ctx context.Context, params acp.RequestPermissionRequest) (acp.RequestPermissionResponse, error) {
	title := ""
	if params.ToolCall.Title != nil {
		title = *params.ToolCall.Title
	}

	if c.AutoApprove {
		log.Printf("[ACP] Auto-approving permission: %s", title)
		for _, o := range params.Options {
			if o.Kind == acp.PermissionOptionKindAllowOnce || o.Kind == acp.PermissionOptionKindAllowAlways {
				return acp.RequestPermissionResponse{
					Outcome: acp.RequestPermissionOutcome{
						Selected: &acp.RequestPermissionOutcomeSelected{OptionId: o.OptionId},
					},
				}, nil
			}
		}
		if len(params.Options) > 0 {
			return acp.RequestPermissionResponse{
				Outcome: acp.RequestPermissionOutcome{
					Selected: &acp.RequestPermissionOutcomeSelected{OptionId: params.Options[0].OptionId},
				},
			}, nil
		}
	}

	log.Printf("[ACP] Permission denied (not auto-approved): %s", title)
	return acp.RequestPermissionResponse{
		Outcome: acp.RequestPermissionOutcome{
			Cancelled: &acp.RequestPermissionOutcomeCancelled{},
		},
	}, nil
}

func (c *MigrationClient) SessionUpdate(ctx context.Context, params acp.SessionNotification) error {
	u := params.Update
	switch {
	case u.AgentMessageChunk != nil:
		content := u.AgentMessageChunk.Content
		if content.Text != nil {
			log.Printf("[Claude] %s", content.Text.Text)
		}
	case u.ToolCall != nil:
		log.Printf("[Tool] %s (%s)", u.ToolCall.Title, u.ToolCall.Status)
	case u.ToolCallUpdate != nil:
		log.Printf("[Tool] %s: %v", u.ToolCallUpdate.ToolCallId, u.ToolCallUpdate.Status)
	case u.AgentThoughtChunk != nil:
		thought := u.AgentThoughtChunk.Content
		if thought.Text != nil {
			log.Printf("[Thought] %s", thought.Text.Text)
		}
	}
	return nil
}

func (c *MigrationClient) WriteTextFile(ctx context.Context, params acp.WriteTextFileRequest) (acp.WriteTextFileResponse, error) {
	path := params.Path
	if !filepath.IsAbs(path) {
		path = filepath.Join(c.WorkDir, path)
	}

	dir := filepath.Dir(path)
	if dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return acp.WriteTextFileResponse{}, fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(path, []byte(params.Content), 0o644); err != nil {
		return acp.WriteTextFileResponse{}, fmt.Errorf("write %s: %w", path, err)
	}
	log.Printf("[ACP] Wrote %d bytes to %s", len(params.Content), path)
	return acp.WriteTextFileResponse{}, nil
}

func (c *MigrationClient) ReadTextFile(ctx context.Context, params acp.ReadTextFileRequest) (acp.ReadTextFileResponse, error) {
	path := params.Path
	if !filepath.IsAbs(path) {
		path = filepath.Join(c.WorkDir, path)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return acp.ReadTextFileResponse{}, fmt.Errorf("read %s: %w", path, err)
	}
	content := string(b)

	if params.Line != nil || params.Limit != nil {
		lines := strings.Split(content, "\n")
		start := 0
		if params.Line != nil && *params.Line > 0 {
			start = min(max(*params.Line-1, 0), len(lines))
		}
		end := len(lines)
		if params.Limit != nil && *params.Limit > 0 {
			if start+*params.Limit < end {
				end = start + *params.Limit
			}
		}
		content = strings.Join(lines[start:end], "\n")
	}

	log.Printf("[ACP] ReadTextFile: %s (%d bytes)", path, len(content))
	return acp.ReadTextFileResponse{Content: content}, nil
}

func (c *MigrationClient) CreateTerminal(ctx context.Context, params acp.CreateTerminalRequest) (acp.CreateTerminalResponse, error) {
	log.Printf("[ACP] CreateTerminal requested (not supported)")
	return acp.CreateTerminalResponse{TerminalId: "term-stub"}, nil
}

func (c *MigrationClient) TerminalOutput(ctx context.Context, params acp.TerminalOutputRequest) (acp.TerminalOutputResponse, error) {
	return acp.TerminalOutputResponse{Output: "", Truncated: false}, nil
}

func (c *MigrationClient) ReleaseTerminal(ctx context.Context, params acp.ReleaseTerminalRequest) (acp.ReleaseTerminalResponse, error) {
	return acp.ReleaseTerminalResponse{}, nil
}

func (c *MigrationClient) WaitForTerminalExit(ctx context.Context, params acp.WaitForTerminalExitRequest) (acp.WaitForTerminalExitResponse, error) {
	return acp.WaitForTerminalExitResponse{}, nil
}

func (c *MigrationClient) KillTerminalCommand(ctx context.Context, params acp.KillTerminalCommandRequest) (acp.KillTerminalCommandResponse, error) {
	return acp.KillTerminalCommandResponse{}, nil
}
