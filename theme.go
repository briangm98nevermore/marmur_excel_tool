package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type scaledTheme struct {
	factor float32
}

func (t *scaledTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, variant)
}

func (t *scaledTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *scaledTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *scaledTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText,
		theme.SizeNameCaptionText,
		theme.SizeNameSubHeadingText,
		theme.SizeNameHeadingText:
		return theme.DefaultTheme().Size(name) * t.factor
	default:
		return theme.DefaultTheme().Size(name)
	}
}