package home

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ilkin0/yaz/internal/device"
)

func initDeviceList(devices []device.Block) list.Model {
	items := make([]list.Item, len(devices))
	for i, d := range devices {
		items[i] = d
	}
	return newCompactList(items, "Device")
}

func (m Model) deviceView() string {
	if m.focused == sectionDevice {
		return focusedStyle.Render(m.devices.View())
	}
	return unfocusedStyle.Render(m.devices.View())
}

func (m *Model) updateDevice(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.devices, cmd = m.devices.Update(msg)
	return cmd
}
