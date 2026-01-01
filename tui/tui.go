package tui

import (
	"fmt"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#888888", Dark: "#666666"}
	highlight = lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#AD8CFF"}
	success   = lipgloss.AdaptiveColor{Light: "#10B981", Dark: "#34D399"}
	warning   = lipgloss.AdaptiveColor{Light: "#F59E0B", Dark: "#FBBF24"}

	titleStyle = lipgloss.NewStyle().
			Foreground(highlight).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(subtle)

	toolStyle = lipgloss.NewStyle().
			Foreground(warning).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(success)

	bufferStyle = lipgloss.NewStyle().
			Foreground(subtle).
			PaddingLeft(2)
)

type ToolExecution struct {
	Name   string
	Status string
	Output []string
}

type Model struct {
	title      string
	status     string
	tools      []ToolExecution
	activeTool int
	buffer     []string
	maxBuffer  int
	mu         sync.Mutex
	done       bool
}

type UpdateMsg struct {
	Type    string
	Title   string
	Status  string
	Content string
}

type DoneMsg struct{}

func New(title string) *Model {
	return &Model{
		title:     title,
		status:    "initializing",
		tools:     []ToolExecution{},
		buffer:    []string{},
		maxBuffer: 8,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case UpdateMsg:
		m.mu.Lock()
		defer m.mu.Unlock()

		switch msg.Type {
		case "status":
			m.status = msg.Status
		case "tool":
			m.tools = append(m.tools, ToolExecution{
				Name:   msg.Title,
				Status: msg.Status,
				Output: []string{},
			})
			m.activeTool = len(m.tools) - 1
		case "tool_update":
			if m.activeTool >= 0 && m.activeTool < len(m.tools) {
				m.tools[m.activeTool].Status = msg.Status
			}
		case "output":
			m.addToBuffer(msg.Content)
		case "thought":
			m.addToBuffer("💭 " + msg.Content)
		}

	case DoneMsg:
		m.mu.Lock()
		m.done = true
		m.status = "completed"
		m.mu.Unlock()
		return m, tea.Quit
	}

	return m, nil
}

func (m *Model) addToBuffer(content string) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			m.buffer = append(m.buffer, line)
			if len(m.buffer) > m.maxBuffer {
				m.buffer = m.buffer[1:]
			}
		}
	}
}

func (m *Model) View() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("◆ "+m.title) + "\n")

	// Status
	statusIcon := "○"
	if m.done {
		statusIcon = "●"
	}
	b.WriteString(statusStyle.Render(fmt.Sprintf("  %s %s", statusIcon, m.status)) + "\n")

	// Tools
	if len(m.tools) > 0 {
		b.WriteString("\n")
		for i, tool := range m.tools {
			icon := "├"
			if i == len(m.tools)-1 {
				icon = "└"
			}
			statusMark := "○"
			if tool.Status == "completed" || tool.Status == "done" {
				statusMark = successStyle.Render("●")
			} else if tool.Status == "running" || tool.Status == "in_progress" {
				statusMark = toolStyle.Render("◐")
			}
			b.WriteString(fmt.Sprintf("  %s %s %s\n", icon, statusMark, tool.Name))
		}
	}

	// Buffer
	if len(m.buffer) > 0 {
		b.WriteString("\n")
		for _, line := range m.buffer {
			if len(line) > 60 {
				line = line[:57] + "..."
			}
			b.WriteString(bufferStyle.Render(line) + "\n")
		}
	}

	return b.String()
}

// SendUpdate sends an update to the TUI
func (m *Model) SendUpdate(program *tea.Program, msgType, title, status, content string) {
	if program != nil {
		program.Send(UpdateMsg{
			Type:    msgType,
			Title:   title,
			Status:  status,
			Content: content,
		})
	}
}

// Complete signals the TUI that the session is done
func (m *Model) Complete(program *tea.Program) {
	if program != nil {
		program.Send(DoneMsg{})
	}
}
