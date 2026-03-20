package home

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/ilkin0/yaz/internal/device"
)

type section int

const (
	sectionDevice section = iota
	sectionImage
	sectionOptions
)

type Model struct {
	focused     section
	devices     list.Model
	source      list.Model
	optionsForm *huh.Form
	opts        Options
	width       int
	height      int
	formWidth   int
}

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
	opts := defaultOptions()
	return Model{
		devices:     initDeviceList(devices),
		source:      initSourceList(),
		optionsForm: initOptionsForm(&opts),
		opts:        opts,
	}
}

func (m Model) Init() tea.Cmd {
	return m.optionsForm.Init()
}

func (m Model) header() string {
	title := titleStyle.Render(fmt.Sprintf(" %s ", appName))
	version := subtitleStyle.Render(fmt.Sprintf("v%s · USB Flasher", appVersion))
	return headerStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Bottom, title, " ", version),
	)
}

func (m Model) helpView() string {
	return helpStyle.Render("tab section · ↑/↓ navigate · enter select · q quit")
}

func (m Model) View() string {
	lists := lipgloss.JoinHorizontal(lipgloss.Top,
		m.deviceView(),
		" ",
		m.imageView(),
		" ",
		m.optionsView(),
	)

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
		sectionWidth := (msg.Width / 3) - 4
		m.formWidth = sectionWidth
		m.devices.SetWidth(innerWidth)
		m.source.SetWidth(innerWidth)
		m.optionsForm.WithWidth(sectionWidth)
		m.devices.SetHeight(listH)
		m.source.SetHeight(listH)

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.focused = (m.focused + 1) % 3
			return m, nil
		case "q":
			if m.focused != sectionOptions {
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	switch m.focused {
	case sectionDevice:
		cmd = m.updateDevice(msg)
	case sectionImage:
		cmd = m.updateImage(msg)
	case sectionOptions:
		cmd = m.updateOptions(msg)
	}
	return m, cmd
}
