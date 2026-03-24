package home

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type imageMode int

const (
	modeSourceList imageMode = iota
	modeBrowse
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
		sourceItem{title: "Enter path manually", desc: "Coming soon", disabled: true},
		sourceItem{title: "Download from URL", desc: "Coming soon", disabled: true},
	}, "Image")
}

func initFilePicker() filepicker.Model {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".iso", ".img", ".raw", ".dmg"}
	fp.DirAllowed = false
	fp.ShowHidden = false
	fp.ShowSize = true
	fp.ShowPermissions = false
	fp.AutoHeight = false
	fp.SetHeight(30)

	home, err := os.UserHomeDir()
	if err == nil {
		fp.CurrentDirectory = home
	} else {
		fp.CurrentDirectory, _ = filepath.Abs(".")
	}

	return fp
}

func (m Model) imageView() string {
	var content string
	switch m.imageMode {
	case modeBrowse:
		content = m.filePicker.View()
	default:
		content = m.source.View()
	}

	if m.focused == sectionImage {
		return focusedStyle.Render(content)
	}
	return unfocusedStyle.Render(content)
}

func (m *Model) updateImage(msg tea.Msg) tea.Cmd {
	switch m.imageMode {
	case modeBrowse:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" {
				m.imageMode = modeSourceList
				return nil
			}
		}

		var cmd tea.Cmd
		m.filePicker, cmd = m.filePicker.Update(msg)

		if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
			m.imagePath = path
			m.imageMode = modeSourceList
			m.source.SetItem(0, sourceItem{
				title: "Browse files...",
				desc:  filepath.Base(path),
			})
		}

		return cmd

	default:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				selected := m.source.SelectedItem()
				if item, ok := selected.(sourceItem); ok {
					if item.disabled {
						return nil
					}
					switch item.title {
					case "Browse files...":
						m.imageMode = modeBrowse
						return m.filePicker.Init()
					}
				}
				return nil
			}
		}

		var cmd tea.Cmd
		m.source, cmd = m.source.Update(msg)
		return cmd
	}
}
