package progress

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ilkin0/yaz/internal/config"
	"github.com/ilkin0/yaz/internal/device"
	"github.com/ilkin0/yaz/internal/writer"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ABABAB"))

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#04B575"))

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF0000"))

	logStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)

// ProgressMsg is sent by the writer goroutine to update the progress bar.
type ProgressMsg writer.Progress

type WriteDoneMsg struct {
	Duration time.Duration
}

type WriteErrorMsg struct {
	Err error
}

type HomeMsg struct{}

type state int

const (
	stateWriting state = iota
	stateDone
	stateError
)

type Model struct {
	program   *tea.Program
	device    device.Block
	image     string
	opts      config.Options
	bar       progress.Model
	state     state
	progress  writer.Progress
	lastPhase writer.Phase
	err       error
	logs      []string
	start     time.Time
	duration  time.Duration
	width     int
	height    int
}

func New(p *tea.Program, dev device.Block, image string, opts config.Options) Model {
	bar := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(50),
	)

	return Model{
		program: p,
		device:  dev,
		image:   image,
		opts:    opts,
		bar:     bar,
		state:   stateWriting,
		start:   time.Now(),
		logs: []string{
			fmt.Sprintf("[%s] Starting write...", time.Now().Format("15:04:05")),
		},
	}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		go func() {
			start := time.Now()
			err := writer.Flash(m.device.DevNode, m.image, m.opts, func(p writer.Progress) {
				m.program.Send(ProgressMsg(p))
			})
			if err != nil {
				m.program.Send(WriteErrorMsg{Err: err})
			} else {
				m.program.Send(WriteDoneMsg{Duration: time.Since(start)})
			}
		}()
		return nil
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.bar.Width = msg.Width / 2

	case ProgressMsg:
		p := writer.Progress(msg)

		if p.LogMessage != "" {
			m.logs = append(m.logs, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), p.LogMessage))
		}

		if p.TotalBytes > 0 {
			m.progress = p
			if m.progress.Phase != m.lastPhase {
				m.lastPhase = m.progress.Phase
			}
			percent := float64(m.progress.BytesWritten) / float64(m.progress.TotalBytes)
			return m, m.bar.SetPercent(percent)
		}

		return m, nil

	case WriteDoneMsg:
		m.state = stateDone
		m.duration = msg.Duration
		m.logs = append(m.logs, fmt.Sprintf("[%s] Write complete!", time.Now().Format("15:04:05")))

	case WriteErrorMsg:
		m.state = stateError
		m.err = msg.Err
		m.logs = append(m.logs, fmt.Sprintf("[%s] ERROR: %v", time.Now().Format("15:04:05"), msg.Err))

	case progress.FrameMsg:
		barModel, cmd := m.bar.Update(msg)
		m.bar = barModel.(progress.Model)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			if m.state == stateDone || m.state == stateError {
				return m, tea.Quit
			}
		case "r":
			if m.state == stateDone || m.state == stateError {
				return m, func() tea.Msg { return HomeMsg{} }
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	phaseLabel := "Writing Image"
	if m.progress.Phase == writer.PhaseVerifying {
		phaseLabel = "Verifying Write"
	}
	title := titleStyle.Render(" " + phaseLabel + " ")

	info := labelStyle.Render(fmt.Sprintf(
		"%s  →  %s",
		filepath.Base(m.image),
		m.device.DevNode,
	))

	var status string
	switch m.state {
	case stateWriting:
		elapsed := time.Since(m.start)
		speed := m.progress.Speed

		eta := ""
		if speed > 0 {
			remaining := float64(m.progress.TotalBytes-m.progress.BytesWritten) / speed
			eta = fmt.Sprintf("ETA: %s", time.Duration(remaining*float64(time.Second)).Truncate(time.Second))
		}

		status = lipgloss.JoinVertical(lipgloss.Left,
			"",
			m.bar.View(),
			"",
			labelStyle.Render(fmt.Sprintf(
				"%s / %s    %s/s    %s    Elapsed: %s",
				humanBytes(m.progress.BytesWritten),
				humanBytes(m.progress.TotalBytes),
				humanBytes(uint64(speed)),
				eta,
				elapsed.Truncate(time.Second),
			)),
		)

	case stateDone:
		avgSpeed := float64(m.progress.TotalBytes) / m.duration.Seconds()
		status = lipgloss.JoinVertical(lipgloss.Left,
			"",
			m.bar.View(),
			"",
			successStyle.Render("Write complete!"),
			labelStyle.Render(fmt.Sprintf(
				"%s written in %s (%s/s)",
				humanBytes(m.progress.TotalBytes),
				m.duration.Truncate(time.Second),
				humanBytes(uint64(avgSpeed)),
			)),
		)

	case stateError:
		status = lipgloss.JoinVertical(lipgloss.Left,
			"",
			errorStyle.Render(fmt.Sprintf("Error: %v", m.err)),
		)
	}

	var sb strings.Builder
	for _, l := range m.logs {
		sb.WriteString(logStyle.Render(l))
		sb.WriteString("\n")
	}
	logContent := sb.String()
	logSection := logStyle.Render("── Log ──") + "\n" + logContent

	help := helpStyle.Render("ctrl+c cancel")
	if m.state == stateDone || m.state == stateError {
		help = helpStyle.Render("q quit · r flash another")
	}

	content := boxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			info,
			status,
			"",
			logSection,
		),
	)

	full := lipgloss.JoinVertical(lipgloss.Center, content, help)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, full)
}

func humanBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
