package style

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type Theme struct {
	// TextColor is the default color of text (typically white)
	TextColor color.Color

	// PanelColor acts as the primary background color to panels and similar elements
	PanelColor color.Color

	// HightlightColor is a varient of PanelColor to be used for higlighting
	HighlightColor color.Color

	// BackgroundColor is the default background of the application
	BackgroundColor color.Color
}

func DefaultTheme() *Theme {
	return &Theme{
		TextColor:       lipgloss.Color("15"),
		PanelColor:      lipgloss.Color("8"),
		HighlightColor:  lipgloss.Color("4"),
		BackgroundColor: lipgloss.Color("0"),
	}
}
