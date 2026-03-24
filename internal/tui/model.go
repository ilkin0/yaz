package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ilkin0/yaz/internal/device"
	"github.com/ilkin0/yaz/internal/tui/screens/confirm"
	"github.com/ilkin0/yaz/internal/tui/screens/home"
	progressScreen "github.com/ilkin0/yaz/internal/tui/screens/progress"
)

type screen int

const (
	screenHome screen = iota
	screenConfirm
	screenProgress
)

type Model struct {
	program *tea.Program

	screen   screen
	home     home.Model
	confirm  confirm.Model
	progress progressScreen.Model

	device *device.Block
	image  string
}

func New(devices []device.Block) *Model {
	return &Model{
		screen: screenHome,
		home:   home.New(devices),
	}
}

func (m Model) Init() tea.Cmd {
	return m.home.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

	case home.ProceedMsg:
		m.confirm = confirm.New(msg.Device, msg.ImagePath, msg.Opts)
		m.screen = screenConfirm
		return m, nil

	case confirm.CancelMsg:
		m.screen = screenHome
		return m, m.home.ResetForm()

	case confirm.ConfirmMsg:
		m.progress = progressScreen.New(m.program, msg.Device, msg.ImagePath, msg.Opts)
		m.screen = screenProgress
		return m, m.progress.Init()
	}

	var cmd tea.Cmd
	var updated tea.Model

	switch m.screen {
	case screenHome:
		updated, cmd = m.home.Update(msg)
		m.home = updated.(home.Model)
	case screenConfirm:
		updated, cmd = m.confirm.Update(msg)
		m.confirm = updated.(confirm.Model)
	case screenProgress:
		updated, cmd = m.progress.Update(msg)
		m.progress = updated.(progressScreen.Model)
	}
	return m, cmd
}

func (m Model) View() string {
	switch m.screen {
	case screenHome:
		return m.home.View()
	case screenConfirm:
		return m.confirm.View()
	case screenProgress:
		return m.progress.View()
	}
	return ""
}
