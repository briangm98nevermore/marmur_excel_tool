package ui

import (
	"fmt"
	"strconv"
	"time"

	"marmur_excel_tool/excel"
	"marmur_excel_tool/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/canvas"
)

type mainUI struct {
	win              fyne.Window
	entries          []model.Entry
	monthSelect      *widget.Select
	yearSelect       *widget.Select
	currencySelect   *widget.Select
	table            *widget.Table
	typeRadio        *widget.RadioGroup
	categorySelect   *widget.Select
	invoiceEntry     *widget.Entry
	descEntry        *widget.Entry
	amountEntry      *widget.Entry
	incomeCategories []string
	expenseCategories []string
}

func BuildUI(w fyne.Window) {
	ui := &mainUI{
		win:              w,
		incomeCategories: append([]string{}, model.DefaultIncomeCategories...),
		expenseCategories: append([]string{}, model.DefaultExpenseCategories...),
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
		func() (int, int) { return len(ui.entries) + 1, 6 },
		func() fyne.CanvasObject {
			// Celda con fondo y texto
			bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
			label := widget.NewLabel("")
			return container.NewStack(bg, label)
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			// obj es un *fyne.Container (Stack)
			stack := obj.(*fyne.Container)
			bg := stack.Objects[0].(*canvas.Rectangle)
			label := stack.Objects[1].(*widget.Label)

			if id.Row == 0 {
				// Cabecera
				headers := []string{"Fecha", "Tipo", "Categoría", "# Factura", "Descripción", "Monto"}
				label.SetText(headers[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
				bg.FillColor = theme.Color(theme.ColorNameHeaderBackground)
				bg.Refresh()
				return
			}

			entry := ui.entries[id.Row-1]
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

			// Color alternado (filas pares e impares)
			if id.Row%2 == 0 {
				bg.FillColor = theme.Color(theme.ColorNameHover)
			} else {
				bg.FillColor = theme.Color(theme.ColorNameBackground)
			}
			bg.Refresh()
		},
	)
	ui.table.SetColumnWidth(0, 100)
	ui.table.SetColumnWidth(1, 80)
	ui.table.SetColumnWidth(2, 120)
	ui.table.SetColumnWidth(3, 80)
	ui.table.SetColumnWidth(4, 200)

	// --- Formulario nueva entrada (orden corregido) ---
	// 1. Primero creamos el selector de categoría
	ui.categorySelect = widget.NewSelect([]string{}, nil)

	// 2. Luego el RadioGroup, cuyo callback ya puede usar categorySelect sin nil pointer
	ui.typeRadio = widget.NewRadioGroup([]string{"Entrada", "Gasto"}, func(selected string) {
		ui.updateCategorySelect()
	})
	ui.typeRadio.Required = true
	// 3. Ahora es seguro seleccionar "Entrada"
	ui.typeRadio.SetSelected("Entrada")

	// Resto de campos
	ui.invoiceEntry = widget.NewEntry()
	ui.invoiceEntry.SetPlaceHolder("# Factura")
	ui.descEntry = widget.NewEntry()
	ui.descEntry.SetPlaceHolder("Descripción")
	ui.amountEntry = widget.NewEntry()
	ui.amountEntry.SetPlaceHolder("Monto (ej. 150.00)")

	btnAdd := widget.NewButton("Añadir", ui.onAddEntry)

	formBox := container.NewVBox(
		widget.NewLabel("Nueva entrada:"),
		container.NewHBox(widget.NewLabel("Tipo:"), ui.typeRadio),
		container.NewHBox(widget.NewLabel("Categoría:"), ui.categorySelect),
		container.NewHBox(widget.NewLabel("# Factura:"), ui.invoiceEntry),
		container.NewHBox(widget.NewLabel("Descripción:"), ui.descEntry),
		container.NewHBox(widget.NewLabel("Monto:"), ui.amountEntry),
		btnAdd,
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

// Cambia las categorías según el tipo seleccionado
func (ui *mainUI) updateCategorySelect() {
	if ui.typeRadio.Selected == "Entrada" {
		ui.categorySelect.Options = ui.incomeCategories
	} else {
		ui.categorySelect.Options = ui.expenseCategories
	}
	ui.categorySelect.Refresh()
}

// Añade una nueva fila
func (ui *mainUI) onAddEntry() {
	// Validar monto
	amount, err := strconv.ParseFloat(ui.amountEntry.Text, 64)
	if err != nil || amount <= 0 {
		dialog.ShowError(fmt.Errorf("Monto inválido"), ui.win)
		return
	}
	// Fecha automática con el mes/año seleccionados
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
	// Si la categoría no existe en la lista, la agregamos
	if ui.typeRadio.Selected == "Entrada" {
		if !contains(ui.incomeCategories, entry.Category) && entry.Category != "" {
			ui.incomeCategories = append(ui.incomeCategories, entry.Category)
		}
	} else {
		if !contains(ui.expenseCategories, entry.Category) && entry.Category != "" {
			ui.expenseCategories = append(ui.expenseCategories, entry.Category)
		}
	}
	ui.entries = append(ui.entries, entry)
	ui.table.Refresh()
	// Limpiar campos
	ui.invoiceEntry.SetText("")
	ui.descEntry.SetText("")
	ui.amountEntry.SetText("")
}

// Generar reporte (llama a la lógica en excel/report.go)
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
		// Llamada a la generación del reporte
		err = excel.GenerateReport(ui.entries, month, year, currency, writer)
		if err != nil {
			dialog.ShowError(err, ui.win)
		} else {
			dialog.ShowInformation("Éxito", "Reporte generado correctamente.", ui.win)
		}
	}, ui.win)
}

// Abre un Excel anterior y carga los datos
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
		// Si el Excel tiene moneda/mes, podrías actualizar los selectores
	}, ui.win)
}

// Guarda los datos actuales en un Excel intermedio
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

// Utilidades
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