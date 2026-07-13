package ui

import (
	"marmur_excel_tool/excel"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func BuildUI(w fyne.Window) {
	// Tabla para mostrar datos del Excel
	table := widget.NewTable(
		func() (int, int) { return 1, 1 },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText("")
		},
	)

	// Función para cargar archivo
	loadFile := func() {
    dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
        if err != nil || reader == nil {
            return
        }
        defer reader.Close()

        path := reader.URI().Path()
        rows, err := excel.ReadAllRows(path, "Sheet1")
        if err != nil {
            dialog.ShowError(err, w)
            return
        }

        // Actualizar tabla dentro de fyne.Do (asegura hilo principal)
        fyne.Do(func() {
            if len(rows) == 0 {
                table.Length = func() (int, int) { return 1, 1 }
                table.Refresh()
                return
            }
            cols := len(rows[0])
            table.Length = func() (int, int) { return len(rows), cols }
            table.UpdateCell = func(id widget.TableCellID, obj fyne.CanvasObject) {
                label := obj.(*widget.Label)
                if id.Row < len(rows) && id.Col < len(rows[id.Row]) {
                    label.SetText(rows[id.Row][id.Col])
                } else {
                    label.SetText("")
                }
            }
            table.Refresh()
        })
    }, w)
}

	btnOpen := widget.NewButton("Abrir Excel", loadFile)
	content := container.NewBorder(nil, btnOpen, nil, nil, table)
	w.SetContent(content)
}