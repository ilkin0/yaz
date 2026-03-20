package home

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type sourceItem struct {
	title    string
	desc     string
	disabled bool
}

func (s sourceItem) Title() string       { return s.title }
func (s sourceItem) Description() string { return s.desc }
func (s sourceItem) FilterValue() string { return s.title }

func initSourceList() list.Model {
	return newCompactList([]list.Item{
		sourceItem{title: "Browse files...", desc: "Select an image from your filesystem"},
		sourceItem{title: "Enter path manually", desc: "Type the full path to an image"},
		sourceItem{title: "Download from URL", desc: "Coming soon", disabled: true},
	}, "Image")
}

func (m Model) imageView() string {
	if m.focused == sectionImage {
		return focusedStyle.Render(m.source.View())
	}
	return unfocusedStyle.Render(m.source.View())
}

func (m *Model) updateImage(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.source, cmd = m.source.Update(msg)
	return cmd
}
