package confirm

import (
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ilkin0/yaz/internal/device"
	"github.com/ilkin0/yaz/internal/tui/screens/home"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	warningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF0000")).
			MarginTop(1).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ABABAB")).
			Width(14)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	sectionStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			MarginBottom(1)

	buttonActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#7D56F4")).
				Padding(0, 2)

	buttonInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ABABAB")).
				Background(lipgloss.Color("#3C3C3C")).
				Padding(0, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)

type ConfirmMsg struct {
	Device    device.Block
	ImagePath string
	Opts      home.Options
}

type CancelMsg struct{}

type Model struct {
	device     device.Block
	imagePath  string
	opts       home.Options
	confirmBtn bool
	width      int
	height     int
}

func New(dev device.Block, imagePath string, opts home.Options) Model {
	return Model{
		device:     dev,
		imagePath:  imagePath,
		opts:       opts,
		confirmBtn: false,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "left", "right", "tab":
			m.confirmBtn = !m.confirmBtn
		case "enter":
			if m.confirmBtn {
				return m, func() tea.Msg {
					return ConfirmMsg{
						Device:    m.device,
						ImagePath: m.imagePath,
						Opts:      m.opts,
					}
				}
			}
			return m, func() tea.Msg { return CancelMsg{} }
		case "esc":
			return m, func() tea.Msg { return CancelMsg{} }
		}
	}

	return m, nil
}

func (m Model) View() string {
	title := titleStyle.Render(" Confirm Flash ")

	warning := warningStyle.Render("WARNING: This will erase ALL data on the target device!")

	sizeGB := fmt.Sprintf("%.1f GB", float64(m.device.BlockSize)/(1024*1024*1024))

	deviceInfo := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render("Device:")+"  "+valueStyle.Render(m.device.DevNode),
		labelStyle.Render("Model:")+"  "+valueStyle.Render(m.device.Model),
		labelStyle.Render("Size:")+"  "+valueStyle.Render(sizeGB),
	)

	imageInfo := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render("Image:")+"  "+valueStyle.Render(filepath.Base(m.imagePath)),
		labelStyle.Render("Path:")+"  "+valueStyle.Render(m.imagePath),
	)

	optsInfo := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render("File System:")+"  "+valueStyle.Render(m.opts.FileSystem),
		labelStyle.Render("Verify:")+"  "+valueStyle.Render(boolToStr(m.opts.VerifyWrite)),
		labelStyle.Render("Sync Mode:")+"  "+valueStyle.Render(boolToStr(m.opts.SyncMode)),
		labelStyle.Render("Quick Format:")+"  "+valueStyle.Render(boolToStr(m.opts.QuickFormat)),
	)

	if m.opts.VolumeLabel != "" {
		optsInfo += "\n" + labelStyle.Render("Label:") + "  " + valueStyle.Render(m.opts.VolumeLabel)
	}

	details := sectionStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			deviceInfo,
			"",
			imageInfo,
			"",
			optsInfo,
		),
	)

	cancelBtn := buttonActiveStyle.Render("Cancel")
	confirmBtn := buttonInactiveStyle.Render("Write")
	if m.confirmBtn {
		cancelBtn = buttonInactiveStyle.Render("Cancel")
		confirmBtn = buttonActiveStyle.Render("Write")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, "   ", confirmBtn)

	help := helpStyle.Render("←/→ switch · enter confirm · esc back")

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		warning,
		details,
		buttons,
		help,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func boolToStr(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
