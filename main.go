package main

import (
	"marmur_excel_tool/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.NewWithID("com.marmur.exceltool")
	w := a.NewWindow("Marmur Excel Tool")

	// Aplicar tema con factor de escala (1.3 = un 30% más grandes)
	a.Settings().SetTheme(&scaledTheme{factor: 1.3})

	ui.BuildUI(w)

	w.Resize(fyne.NewSize(1200, 700))
	w.ShowAndRun()
}