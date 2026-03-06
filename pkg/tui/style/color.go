package style

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type Theme struct {
	TextColor             color.Color
	BackgroundColor       color.Color
	BackgroundAccentColor color.Color
	PrimaryColor          color.Color
	AccentColor           color.Color
}

func DefaultTheme() *Theme {
	return &Theme{
		TextColor:             lipgloss.Color("15"),
		BackgroundColor:       lipgloss.Color("0"),
		BackgroundAccentColor: lipgloss.Color("8"),
		PrimaryColor:          lipgloss.Color("4"),
		AccentColor:           lipgloss.Color("14"),
	}
}
