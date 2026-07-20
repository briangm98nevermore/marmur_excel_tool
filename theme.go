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
	switch name {
	// Fondo principal: gris muy claro
	case theme.ColorNameBackground:
		return color.NRGBA{R: 232, G: 232, B: 232, A: 255} // #E8E8E8

	// Cabeceras de tabla: un poco más oscuro que el fondo
	case theme.ColorNameHeaderBackground:
		return color.NRGBA{R: 216, G: 216, B: 216, A: 255} // #D8D8D8

	// Filas alternas / hover: más oscuro para contraste
	case theme.ColorNameHover:
		return color.NRGBA{R: 208, G: 208, B: 208, A: 255} // #D0D0D0

	// Botones y campos de entrada: casi blancos pero suaves
	case theme.ColorNameButton, theme.ColorNameInputBackground:
		return color.NRGBA{R: 245, G: 245, B: 245, A: 255} // #F5F5F5

	// Resto de colores: mantener el tema claro por defecto
	default:
		return theme.DefaultTheme().Color(name, theme.VariantLight)
	}
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