package style

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	TextColor       lipgloss.Color
	BackgroundColor lipgloss.Color
	PrimaryColor    lipgloss.Color
	AccentColor     lipgloss.Color
}

func DefaultTheme() *Theme {
	return &Theme{
		TextColor:       lipgloss.Color("15"),
		BackgroundColor: lipgloss.Color("0"),
		PrimaryColor:    lipgloss.Color("4"),
		AccentColor:     lipgloss.Color("14"),
	}
}
