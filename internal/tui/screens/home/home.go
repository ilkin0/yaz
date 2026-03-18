package home

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ilkin0/yaz/internal/device"
)

var (
	appName    = "yaz"
	appVersion = "0.1.0"

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ABABAB")).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			MarginBottom(1).
			Padding(1, 2)
)

type Model struct {
	devices list.Model
}

func New(devices []device.Block) Model {
	items := make([]list.Item, len(devices))
	for i, d := range devices {
		items[i] = d
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a device"
	l.SetShowHelp(true)
	l.SetFilteringEnabled(false)

	return Model{devices: l}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) header() string {
	title := titleStyle.Render(fmt.Sprintf(" %s ", appName))
	version := subtitleStyle.Render(fmt.Sprintf("v%s · USB Flasher", appVersion))
	return headerStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, title, version),
	)
}

func (m Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.header(),
		m.devices.View(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.devices.SetWidth(msg.Width)
		headerHeight := lipgloss.Height(m.header())
		m.devices.SetHeight(msg.Height - headerHeight)
	}

	var cmd tea.Cmd
	m.devices, cmd = m.devices.Update(msg)
	return m, cmd
}
