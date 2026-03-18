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
			Foreground(lipgloss.Color("#ABABAB"))

	headerStyle = lipgloss.NewStyle().
			Padding(1, 2).
			MarginBottom(0)

	focusedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(0)

	unfocusedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3C3C3C")).
			Padding(0, 1).
			MarginBottom(0)
)

type Model struct {
	focused section
	devices list.Model
	source  list.Model
	width   int
	height  int
}

type sourceItem struct {
	title    string
	desc     string
	disabled bool
}

func (s sourceItem) Title() string       { return s.title }
func (s sourceItem) Description() string { return s.desc }
func (s sourceItem) FilterValue() string { return s.title }

type section int

const (
	sectionDevice section = iota
	sectionImage
)

func newCompactList(items []list.Item, title string) list.Model {
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = title
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)
	return l
}

func New(devices []device.Block) Model {
	items := make([]list.Item, len(devices))
	for i, d := range devices {
		items[i] = d
	}

	deviceList := newCompactList(items, "Device")

	sourceList := newCompactList([]list.Item{
		sourceItem{title: "Browse files...", desc: "Select an image from your filesystem"},
		sourceItem{title: "Enter path manually", desc: "Type the full path to an image"},
		sourceItem{title: "Download from URL", desc: "Coming soon", disabled: true},
	}, "Image")

	return Model{devices: deviceList, source: sourceList}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) header() string {
	title := titleStyle.Render(fmt.Sprintf(" %s ", appName))
	version := subtitleStyle.Render(fmt.Sprintf("v%s · USB Flasher", appVersion))
	return headerStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Bottom, title, " ", version),
	)
}

func (m Model) helpView() string {
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Padding(0, 2).
		Render("tab section · ↑/↓ navigate · enter select · q quit")
	return help
}

func (m Model) View() string {
	deviceView := unfocusedStyle.Render(m.devices.View())
	sourceView := unfocusedStyle.Render(m.source.View())

	if m.focused == sectionDevice {
		deviceView = focusedStyle.Render(m.devices.View())
	} else {
		sourceView = focusedStyle.Render(m.source.View())
	}

	lists := lipgloss.JoinHorizontal(lipgloss.Top, deviceView, " ", sourceView)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		m.header(),
		lists,
		m.helpView(),
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, content)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerH := lipgloss.Height(m.header())
		helpH := lipgloss.Height(m.helpView())
		borderPadding := 6
		available := msg.Height - headerH - helpH - borderPadding
		listH := available / 2

		innerWidth := msg.Width - 4
		m.devices.SetWidth(innerWidth)
		m.source.SetWidth(innerWidth)
		m.devices.SetHeight(listH)
		m.source.SetHeight(listH)

	case tea.KeyMsg:
		if msg.String() == "tab" {
			if m.focused == sectionDevice {
				m.focused = sectionImage
			} else {
				m.focused = sectionDevice
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	switch m.focused {
	case sectionDevice:
		m.devices, cmd = m.devices.Update(msg)
	case sectionImage:
		m.source, cmd = m.source.Update(msg)
	}
	return m, cmd
}
