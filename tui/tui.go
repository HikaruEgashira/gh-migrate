package tui

import (
	"fmt"
	"os"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#888888", Dark: "#666666"}
	highlight = lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#AD8CFF"}
	success   = lipgloss.AdaptiveColor{Light: "#10B981", Dark: "#34D399"}
	warning   = lipgloss.AdaptiveColor{Light: "#F59E0B", Dark: "#FBBF24"}
	errColor  = lipgloss.AdaptiveColor{Light: "#EF4444", Dark: "#F87171"}

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

	errorStyle = lipgloss.NewStyle().
			Foreground(errColor)

	bufferStyle = lipgloss.NewStyle().
			Foreground(subtle).
			PaddingLeft(2)
)

type Step struct {
	Name   string
	Status string // pending, running, done, error
}

type Model struct {
	title        string
	status       string
	steps        []Step
	buffer       []string
	maxBuffer    int
	mu           sync.Mutex
	done         bool
	lineComplete bool // 前の行が改行で終わったかどうか
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
		steps:     []Step{},
		buffer:    []string{},
		maxBuffer: 6,
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
		case "step":
			m.steps = append(m.steps, Step{Name: msg.Title, Status: msg.Status})
		case "step_update":
			if len(m.steps) > 0 {
				m.steps[len(m.steps)-1].Status = msg.Status
			}
		case "tool":
			m.steps = append(m.steps, Step{Name: msg.Title, Status: msg.Status})
		case "tool_update":
			if len(m.steps) > 0 {
				m.steps[len(m.steps)-1].Status = msg.Status
			}
		case "output", "log":
			m.addToBuffer(msg.Content)
		case "thought":
			m.addToBuffer("💭 " + msg.Content)
		case "error":
			m.addToBuffer("✗ " + msg.Content)
		case "success":
			m.addToBuffer("✓ " + msg.Content)
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
	if content == "" {
		return
	}

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 前の行が完了していない場合、最後のバッファに追加
		if i == 0 && len(m.buffer) > 0 && !m.lineComplete {
			m.buffer[len(m.buffer)-1] += " " + line
		} else {
			m.buffer = append(m.buffer, line)
			if len(m.buffer) > m.maxBuffer {
				m.buffer = m.buffer[1:]
			}
		}
	}

	// 改行で終わっているかどうかを追跡
	m.lineComplete = strings.HasSuffix(content, "\n")
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

	// Steps
	if len(m.steps) > 0 {
		b.WriteString("\n")
		for i, step := range m.steps {
			icon := "├"
			if i == len(m.steps)-1 {
				icon = "└"
			}
			statusMark := "○"
			switch step.Status {
			case "done", "completed":
				statusMark = successStyle.Render("●")
			case "running", "in_progress":
				statusMark = toolStyle.Render("◐")
			case "error":
				statusMark = errorStyle.Render("✗")
			}
			b.WriteString(fmt.Sprintf("  %s %s %s\n", icon, statusMark, step.Name))
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

// UI is a global TUI manager
type UI struct {
	model   *Model
	program *tea.Program
	isTTY   bool
	mu      sync.Mutex
}

var globalUI *UI

// Init initializes the global UI
func Init(title string) *UI {
	isTTY := term.IsTerminal(int(os.Stderr.Fd()))

	ui := &UI{
		isTTY: isTTY,
	}

	if isTTY {
		ui.model = New(title)
		ui.program = tea.NewProgram(ui.model, tea.WithOutput(os.Stderr))

		go func() {
			ui.program.Run()
		}()
	}

	globalUI = ui
	return ui
}

// Get returns the global UI
func Get() *UI {
	return globalUI
}

// Status updates the status
func (u *UI) Status(status string) {
	if u == nil {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.isTTY && u.model != nil {
		u.model.SendUpdate(u.program, "status", "", status, "")
	} else {
		fmt.Fprintf(os.Stderr, "○ %s\n", status)
	}
}

// Step adds a new step
func (u *UI) Step(name string) {
	if u == nil {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.isTTY && u.model != nil {
		u.model.SendUpdate(u.program, "step", name, "running", "")
	} else {
		fmt.Fprintf(os.Stderr, "◐ %s\n", name)
	}
}

// StepDone marks the current step as done
func (u *UI) StepDone() {
	if u == nil {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.isTTY && u.model != nil {
		u.model.SendUpdate(u.program, "step_update", "", "done", "")
	}
}

// StepError marks the current step as error
func (u *UI) StepError() {
	if u == nil {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.isTTY && u.model != nil {
		u.model.SendUpdate(u.program, "step_update", "", "error", "")
	}
}

// Log adds a log message
func (u *UI) Log(format string, args ...interface{}) {
	if u == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)

	u.mu.Lock()
	defer u.mu.Unlock()

	if u.isTTY && u.model != nil {
		u.model.SendUpdate(u.program, "log", "", "", msg)
	} else {
		fmt.Fprintf(os.Stderr, "  %s\n", msg)
	}
}

// Success logs a success message
func (u *UI) Success(format string, args ...interface{}) {
	if u == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)

	u.mu.Lock()
	defer u.mu.Unlock()

	if u.isTTY && u.model != nil {
		u.model.SendUpdate(u.program, "success", "", "", msg)
	} else {
		fmt.Fprintf(os.Stderr, "✓ %s\n", msg)
	}
}

// Error logs an error message
func (u *UI) Error(format string, args ...interface{}) {
	if u == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)

	u.mu.Lock()
	defer u.mu.Unlock()

	if u.isTTY && u.model != nil {
		u.model.SendUpdate(u.program, "error", "", "", msg)
	} else {
		fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
	}
}

// Done completes the UI
func (u *UI) Done() {
	if u == nil {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.isTTY && u.model != nil {
		u.model.Complete(u.program)
	}
}

// GetModel returns the model (for ACP integration)
func (u *UI) GetModel() *Model {
	if u == nil {
		return nil
	}
	return u.model
}

// GetProgram returns the program (for ACP integration)
func (u *UI) GetProgram() *tea.Program {
	if u == nil {
		return nil
	}
	return u.program
}

// IsTTY returns whether the UI is running in a TTY
func (u *UI) IsTTY() bool {
	if u == nil {
		return false
	}
	return u.isTTY
}
