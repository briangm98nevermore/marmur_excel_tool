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

	// --- Tabla con edición/eliminación y colores alternados ---
	ui.table = widget.NewTable(
		func() (int, int) { return len(ui.entries) + 1, 8 },
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
			content := container.NewHBox()
			return container.NewStack(bg, content)
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			stack := obj.(*fyne.Container)
			bg := stack.Objects[0].(*canvas.Rectangle)
			// Limpiar contenido anterior
			if len(stack.Objects) > 1 {
				stack.Remove(stack.Objects[1])
			}
			var content fyne.CanvasObject

			if id.Row == 0 {
				headers := []string{"Fecha", "Tipo", "Categoría", "# Factura", "Descripción", "Monto", "Editar", "Eliminar"}
				label := widget.NewLabel(headers[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
				content = label
				bg.FillColor = theme.Color(theme.ColorNameHeaderBackground)
				bg.Refresh()
			} else {
				entry := ui.entries[id.Row-1]
				if id.Col < 6 {
					label := widget.NewLabel("")
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
					}
					content = label
				} else if id.Col == 6 {
					btn := widget.NewButton("Editar", func() {
						ui.startEdit(id.Row - 1)
					})
					content = btn
				} else if id.Col == 7 {
					btn := widget.NewButton("Eliminar", func() {
						ui.deleteEntry(id.Row - 1)
					})
					content = btn
				}
				// Color alternado de filas
				if id.Row%2 == 0 {
					bg.FillColor = theme.Color(theme.ColorNameHover)
				} else {
					bg.FillColor = theme.Color(theme.ColorNameBackground)
				}
				bg.Refresh()
			}
			stack.Add(content)
		},
	)
	ui.table.SetColumnWidth(0, 100)
	ui.table.SetColumnWidth(1, 80)
	ui.table.SetColumnWidth(2, 120)
	ui.table.SetColumnWidth(3, 100)
	ui.table.SetColumnWidth(4, 200)
	ui.table.SetColumnWidth(5, 100)
	ui.table.SetColumnWidth(6, 80)
	ui.table.SetColumnWidth(7, 80)

	// --- Formulario (usa widget.NewForm para campos más grandes) ---
	ui.categorySelect = widget.NewSelect([]string{}, nil)

	ui.typeRadio = widget.NewRadioGroup([]string{"Entrada", "Gasto"}, func(selected string) {
		ui.updateCategorySelect()
		// Seleccionar automáticamente primera categoría al cambiar tipo
		if selected == "Entrada" && len(ui.incomeCategories) > 0 {
			ui.categorySelect.SetSelected(ui.incomeCategories[0])
		} else if selected == "Gasto" && len(ui.expenseCategories) > 0 {
			ui.categorySelect.SetSelected(ui.expenseCategories[0])
		}
	})
	ui.typeRadio.Required = true
	ui.typeRadio.SetSelected("Entrada")

	ui.invoiceEntry = widget.NewEntry()
	ui.invoiceEntry.SetPlaceHolder("# Factura")
	ui.descEntry = widget.NewEntry()
	ui.descEntry.SetPlaceHolder("Descripción")
	ui.amountEntry = widget.NewEntry()
	ui.amountEntry.SetPlaceHolder("Monto (ej. 150.00)")

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

// Actualiza las opciones del selector de categoría según el tipo elegido
func (ui *mainUI) updateCategorySelect() {
	if ui.typeRadio.Selected == "Entrada" {
		ui.categorySelect.Options = ui.incomeCategories
	} else {
		ui.categorySelect.Options = ui.expenseCategories
	}
	ui.categorySelect.Refresh()
}

// Añade o actualiza una entrada
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
		// Actualizar entrada existente
		ui.entries[ui.editingIndex] = entry
		ui.editingIndex = -1
		ui.addButton.SetText("Añadir")
		ui.cancelButton.Hide()
	} else {
		// Nueva entrada: agregar categoría si no existe
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

	// Limpiar formulario
	ui.invoiceEntry.SetText("")
	ui.descEntry.SetText("")
	ui.amountEntry.SetText("")
	ui.table.Refresh()
}

// Carga los datos de una entrada en el formulario para editar
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

// Elimina una entrada con confirmación
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

// Cancela la edición actual
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

	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
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
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			return
		}
		defer writer.Close()
		err = excel.SaveEntries(ui.entries, writer)
		if err != nil {
			dialog.ShowError(err, ui.win)
		}
	}, ui.win)
}

// Funciones auxiliares
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