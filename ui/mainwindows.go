package ui

import (
	"fmt"
	"strconv"
	"time"

	"marmur_excel_tool/excel"
	"marmur_excel_tool/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type mainUI struct {
	win               fyne.Window
	entries           []model.Entry
	monthSelect       *widget.Select
	yearSelect        *widget.Select
	currencySelect    *widget.Select
	table             *widget.Table
	typeRadio         *widget.RadioGroup
	categorySelect    *widget.Select
	invoiceEntry      *widget.Entry
	descEntry         *widget.Entry
	amountEntry       *widget.Entry
	incomeCategories  []string
	expenseCategories []string
	editingIndex      int
	addButton         *widget.Button
	cancelButton      *widget.Button
}

func BuildUI(w fyne.Window) {
	ui := &mainUI{
		win:              w,
		incomeCategories: append([]string{}, model.DefaultIncomeCategories...),
		expenseCategories: append([]string{}, model.DefaultExpenseCategories...),
		editingIndex:     -1,
	}

	// --- Selectores superiores ---
	months := []string{"Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio",
		"Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre"}
	currentMonth := months[time.Now().Month()-1]
	ui.monthSelect = widget.NewSelect(months, nil)
	ui.monthSelect.SetSelected(currentMonth)

	years := []string{"2024", "2025", "2026", "2027"}
	ui.yearSelect = widget.NewSelect(years, nil)
	ui.yearSelect.SetSelected(strconv.Itoa(time.Now().Year()))

	ui.currencySelect = widget.NewSelect([]string{"USD", "CUP", "EUR"}, nil)
	ui.currencySelect.SetSelected("USD")

	btnReport := widget.NewButton("Generar Reporte", ui.onGenerateReport)
	btnReport.Importance = widget.HighImportance

	btnOpen := widget.NewButton("Abrir Excel", ui.onOpenExcel)
	btnSave := widget.NewButton("Guardar", ui.onSave)

	topBar := container.NewHBox(
		widget.NewLabel("Mes:"), ui.monthSelect,
		widget.NewLabel("Año:"), ui.yearSelect,
		widget.NewLabel("Moneda:"), ui.currencySelect,
		btnReport,
		btnOpen,
		btnSave,
	)

	// --- Tabla ---
	ui.table = widget.NewTable(
		func() (int, int) { 
			return len(ui.entries) + 1, 8 
		},
		func() fyne.CanvasObject {
			// Cada celda tendrá un fondo y una etiqueta dentro de un Stack
			bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
			label := widget.NewLabel("")
			label.Wrapping = fyne.TextTruncate
			stack := container.NewStack(bg, label)
			return stack
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			stack := obj.(*fyne.Container)
			bg := stack.Objects[0].(*canvas.Rectangle)
			label := stack.Objects[1].(*widget.Label)

			if id.Row == 0 {
				// Cabecera
				headers := []string{"Fecha", "Tipo", "Categoría", "# Factura", "Descripción", "Monto", "Editar", "Eliminar"}
				label.SetText(headers[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
				bg.FillColor = theme.Color(theme.ColorNameHeaderBackground)
				bg.Refresh()
				return
			}

			entry := ui.entries[id.Row-1]
			
			// Datos de la fila
			switch id.Col {
			case 0:
				label.SetText(entry.Date.Format("02/01/2006"))
			case 1:
				label.SetText(entry.Type)
			case 2:
				label.SetText(entry.Category)
			case 3:
				label.SetText(entry.InvoiceNum)
			case 4:
				label.SetText(entry.Description)
			case 5:
				label.SetText(fmt.Sprintf("%.2f %s", entry.Amount, entry.Currency))
			case 6:
				label.SetText("✏️ Editar")
				label.TextStyle = fyne.TextStyle{Italic: true}
			case 7:
				label.SetText("🗑️ Eliminar")
				label.TextStyle = fyne.TextStyle{Italic: true}
			}

			// Color alternado de filas
			if id.Row%2 == 0 {
				bg.FillColor = theme.Color(theme.ColorNameHover)
			} else {
				bg.FillColor = theme.Color(theme.ColorNameBackground)
			}
			bg.Refresh()
		},
	)
	
	// Hacer clic en las celdas de editar/eliminar
	ui.table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			return // Cabecera
		}
		index := id.Row - 1
		switch id.Col {
		case 6:
			ui.startEdit(index)
		case 7:
			ui.deleteEntry(index)
		}
	}

	ui.table.SetColumnWidth(0, 100)
	ui.table.SetColumnWidth(1, 80)
	ui.table.SetColumnWidth(2, 120)
	ui.table.SetColumnWidth(3, 100)
	ui.table.SetColumnWidth(4, 200)
	ui.table.SetColumnWidth(5, 100)
	ui.table.SetColumnWidth(6, 100)
	ui.table.SetColumnWidth(7, 100)

	// --- Formulario ---
	ui.categorySelect = widget.NewSelect([]string{}, nil)

	ui.typeRadio = widget.NewRadioGroup([]string{"Entrada", "Gasto"}, func(selected string) {
		ui.updateCategorySelect()
		if selected == "Entrada" && len(ui.incomeCategories) > 0 {
			ui.categorySelect.SetSelected(ui.incomeCategories[0])
		} else if selected == "Gasto" && len(ui.expenseCategories) > 0 {
			ui.categorySelect.SetSelected(ui.expenseCategories[0])
		}
	})
	ui.typeRadio.Required = true
	ui.typeRadio.SetSelected("Entrada")

	// Campos de entrada con tamaño adecuado
	ui.invoiceEntry = widget.NewEntry()
	ui.invoiceEntry.SetPlaceHolder("# Factura")
	ui.invoiceEntry.SetMinRowsVisible(1)
	
	ui.descEntry = widget.NewEntry()
	ui.descEntry.SetPlaceHolder("Descripción")
	ui.descEntry.SetMinRowsVisible(1)
	
	ui.amountEntry = widget.NewEntry()
	ui.amountEntry.SetPlaceHolder("Monto (ej. 150.00)")
	ui.amountEntry.SetMinRowsVisible(1)

	form := widget.NewForm(
		widget.NewFormItem("Tipo", ui.typeRadio),
		widget.NewFormItem("Categoría", ui.categorySelect),
		widget.NewFormItem("# Factura", ui.invoiceEntry),
		widget.NewFormItem("Descripción", ui.descEntry),
		widget.NewFormItem("Monto", ui.amountEntry),
	)

	ui.addButton = widget.NewButton("Añadir", ui.onAddOrUpdateEntry)
	ui.cancelButton = widget.NewButton("Cancelar", ui.onCancelEdit)
	ui.cancelButton.Hide()
	btnBox := container.NewHBox(ui.addButton, ui.cancelButton)

	formBox := container.NewVBox(
		widget.NewLabel("Nueva entrada:"),
		form,
		btnBox,
	)

	// Layout final
	content := container.NewBorder(
		topBar,
		formBox,
		nil, nil,
		ui.table,
	)
	w.SetContent(content)
}

func (ui *mainUI) updateCategorySelect() {
	if ui.typeRadio.Selected == "Entrada" {
		ui.categorySelect.Options = ui.incomeCategories
	} else {
		ui.categorySelect.Options = ui.expenseCategories
	}
	ui.categorySelect.Refresh()
}

func (ui *mainUI) onAddOrUpdateEntry() {
	amount, err := strconv.ParseFloat(ui.amountEntry.Text, 64)
	if err != nil || amount <= 0 {
		dialog.ShowError(fmt.Errorf("Monto inválido"), ui.win)
		return
	}
	year, _ := strconv.Atoi(ui.yearSelect.Selected)
	month := indexOfMonth(ui.monthSelect.Selected) + 1
	date := time.Date(year, time.Month(month), time.Now().Day(), 0, 0, 0, 0, time.Local)

	entry := model.Entry{
		Date:        date,
		Type:        ui.typeRadio.Selected,
		Category:    ui.categorySelect.Selected,
		InvoiceNum:  ui.invoiceEntry.Text,
		Description: ui.descEntry.Text,
		Amount:      amount,
		Currency:    ui.currencySelect.Selected,
	}

	if ui.editingIndex >= 0 {
		ui.entries[ui.editingIndex] = entry
		ui.editingIndex = -1
		ui.addButton.SetText("Añadir")
		ui.cancelButton.Hide()
	} else {
		if entry.Type == "Entrada" {
			if !contains(ui.incomeCategories, entry.Category) && entry.Category != "" {
				ui.incomeCategories = append(ui.incomeCategories, entry.Category)
			}
		} else {
			if !contains(ui.expenseCategories, entry.Category) && entry.Category != "" {
				ui.expenseCategories = append(ui.expenseCategories, entry.Category)
			}
		}
		ui.entries = append(ui.entries, entry)
	}

	ui.invoiceEntry.SetText("")
	ui.descEntry.SetText("")
	ui.amountEntry.SetText("")
	ui.table.Refresh()
}

func (ui *mainUI) startEdit(index int) {
	if index < 0 || index >= len(ui.entries) {
		return
	}
	entry := ui.entries[index]
	ui.typeRadio.SetSelected(entry.Type)
	ui.updateCategorySelect()
	ui.categorySelect.SetSelected(entry.Category)
	ui.invoiceEntry.SetText(entry.InvoiceNum)
	ui.descEntry.SetText(entry.Description)
	ui.amountEntry.SetText(fmt.Sprintf("%.2f", entry.Amount))
	ui.editingIndex = index
	ui.addButton.SetText("Actualizar")
	ui.cancelButton.Show()
}

func (ui *mainUI) deleteEntry(index int) {
	if index < 0 || index >= len(ui.entries) {
		return
	}
	dialog.ShowConfirm("Eliminar", "¿Eliminar esta entrada?", func(ok bool) {
		if ok {
			ui.entries = append(ui.entries[:index], ui.entries[index+1:]...)
			ui.table.Refresh()
		}
	}, ui.win)
}

func (ui *mainUI) onCancelEdit() {
	ui.editingIndex = -1
	ui.addButton.SetText("Añadir")
	ui.cancelButton.Hide()
	ui.invoiceEntry.SetText("")
	ui.descEntry.SetText("")
	ui.amountEntry.SetText("")
}

// Genera el reporte Excel
func (ui *mainUI) onGenerateReport() {
	if len(ui.entries) == 0 {
		dialog.ShowInformation("Sin datos", "No hay entradas para generar el reporte.", ui.win)
		return
	}
	month := ui.monthSelect.Selected
	year := ui.yearSelect.Selected
	currency := ui.currencySelect.Selected
	
	dialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			return
		}
		defer writer.Close()
		err = excel.GenerateReport(ui.entries, month, year, currency, writer)
		if err != nil {
			dialog.ShowError(err, ui.win)
		} else {
			dialog.ShowInformation("Éxito", "Reporte generado correctamente.", ui.win)
		}
	}, ui.win)
	
	fileName := fmt.Sprintf("reporte_%s_%s_%s.xlsx", month, year, currency)
	dialog.SetFileName(fileName)
	dialog.Show()
}

// Abre un Excel existente para cargar entradas
func (ui *mainUI) onOpenExcel() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()
		path := reader.URI().Path()
		entries, err := excel.ImportEntries(path)
		if err != nil {
			dialog.ShowError(err, ui.win)
			return
		}
		ui.entries = entries
		ui.table.Refresh()
	}, ui.win)
}

// Guarda las entradas actuales en un Excel (para continuar después)
func (ui *mainUI) onSave() {
	dialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			return
		}
		defer writer.Close()
		err = excel.SaveEntries(ui.entries, writer)
		if err != nil {
			dialog.ShowError(err, ui.win)
		} else {
			dialog.ShowInformation("Guardado", "Datos guardados correctamente.", ui.win)
		}
	}, ui.win)
	
	dialog.SetFileName("marmur_datos.xlsx")
	dialog.Show()
}

func indexOfMonth(month string) int {
	months := []string{"Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio",
		"Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre"}
	for i, m := range months {
		if m == month {
			return i
		}
	}
	return 0
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}