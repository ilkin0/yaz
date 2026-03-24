package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ilkin0/yaz/internal/device"
)

func Start() {
	devices, err := device.EnumurateDevice()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	m := New(devices)
	p := tea.NewProgram(m, tea.WithAltScreen())
	m.program = p
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
