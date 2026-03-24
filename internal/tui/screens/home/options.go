package home

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/ilkin0/yaz/internal/config"
)

func defaultOptions() config.Options {
	return config.Options{
		VerifyWrite: true,
		QuickFormat: true,
		FileSystem:  "fat32",
		ClusterSize: "auto",
	}
}

func initOptionsForm(opts *config.Options) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Quick Format").
				Description("Skip zero-fill before writing").
				Value(&opts.QuickFormat),

			huh.NewConfirm().
				Title("Verify Write").
				Description("Compare checksums after write").
				Value(&opts.VerifyWrite),

			huh.NewConfirm().
				Title("Sync Mode").
				Description("Use O_SYNC for safer writes").
				Value(&opts.SyncMode),
		),

		// Phase 2: format-only mode options (disabled)
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("File System").
				Description("Coming in Phase 2").
				Options(
					huh.NewOption("FAT32", "fat32"),
					huh.NewOption("exFAT", "exfat"),
					huh.NewOption("ext4", "ext4"),
				).
				Value(&opts.FileSystem),

			huh.NewSelect[string]().
				Title("Cluster Size").
				Description("Coming in Phase 2").
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
				Description("Coming in Phase 2").
				Value(&opts.VolumeLabel),
		).WithHide(true),
	)
}

func (m Model) optionsView() string {
	title := optionsTitleStyle.Render("Write Options")
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
