package home

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type Options struct {
	VerifyWrite bool
	SyncMode    bool
	QuickFormat bool
	FileSystem  string
	VolumeLabel string
	ClusterSize string
}

func defaultOptions() Options {
	return Options{
		VerifyWrite: true,
		FileSystem:  "fat32",
		ClusterSize: "auto",
	}
}

func initOptionsForm(opts *Options) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Verify Write").
				Value(&opts.VerifyWrite),

			huh.NewConfirm().
				Title("Sync Mode").
				Value(&opts.SyncMode),

			huh.NewConfirm().
				Title("Quick Format").
				Value(&opts.QuickFormat),

			huh.NewSelect[string]().
				Title("File System").
				Options(
					huh.NewOption("FAT32", "fat32"),
					huh.NewOption("exFAT", "exfat"),
					huh.NewOption("ext4", "ext4"),
				).
				Value(&opts.FileSystem),

			huh.NewSelect[string]().
				Title("Cluster Size").
				Options(
					huh.NewOption("Auto", "auto"),
					huh.NewOption("4K", "4k"),
					huh.NewOption("8K", "8k"),
					huh.NewOption("16K", "16k"),
					huh.NewOption("32K", "32k"),
					huh.NewOption("64K", "64k"),
				).
				Value(&opts.ClusterSize),

			huh.NewInput().
				Title("Volume Label").
				Value(&opts.VolumeLabel),
		),
	)
}

func (m Model) optionsView() string {
	title := optionsTitleStyle.Render("Format Options")
	content := title + "\n" + m.optionsForm.View()

	if m.focused == sectionOptions {
		return focusedStyle.Render(content)
	}
	return unfocusedStyle.Render(content)
}

func (m *Model) updateOptions(msg tea.Msg) tea.Cmd {
	form, cmd := m.optionsForm.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.optionsForm = f
	}

	if m.optionsForm.State == huh.StateCompleted {
		dev, ok := m.SelectedDevice()
		if ok && m.imagePath != "" {
			return func() tea.Msg {
				return ProceedMsg{
					Device:    dev,
					ImagePath: m.imagePath,
					Opts:      *m.opts,
				}
			}
		}
	}

	return cmd
}
