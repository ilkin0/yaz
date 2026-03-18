package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ilkin0/yaz/internal/device"
	"github.com/ilkin0/yaz/internal/tui/screens/home"
)

type screen int

const (
	screenHome screen = iota
	screenImageSelect
	screenConfirm
	screenProgress
	screenComplete
)

type Model struct {
	screen screen
	home   home.Model

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
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var updated tea.Model

	switch m.screen {
	case screenHome:
		updated, cmd = m.home.Update(msg)
		m.home = updated.(home.Model)
	}
	return m, cmd
}

func (m Model) View() string {
	switch m.screen {
	case screenHome:
		return m.home.View()
	}
	return ""
}
