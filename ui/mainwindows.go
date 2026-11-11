package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"sort"

	"marmur_excel_tool/excel"
	"marmur_excel_tool/model"

	"image/color"

	xwidget "fyne.io/x/fyne/widget"


	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Meses del año (global para poder usarla en onClear)
var months = []string{
	"Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio",
	"Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre",
}

type mainUI struct {
	app               fyne.App
	win               fyne.Window
	entries           []model.Entry
	monthSelect       *widget.Select
	yearSelect        *widget.Select
	currencySelect    *widget.Select
	incomeTable    	  *widget.Table
	expenseTable   	  *widget.Table
	typeRadio         *widget.RadioGroup
	categorySelect    *widget.Select
	invoiceEntry      *widget.Entry
	descEntry 		  *xwidget.CompletionEntry
	amountEntry       *widget.Entry
	dayEntry          *widget.Entry
	incomeCategories  []string
	expenseCategories []string
	editingIndex      int
	addButton         *widget.Button
	cancelButton      *widget.Button
	displayEntries         []model.Entry // copia ordenada para mostrar
    displayOriginalIndices []int          // índice en ui.entries correspondiente a cada fila mostrada
	displayIncomeEntries  []model.Entry
	displayIncomeIndices  []int
	displayExpenseEntries []model.Entry
	displayExpenseIndices []int
	allDescriptions []string
}

type indexedEntry struct {
    entry model.Entry
    idx   int
}

func BuildUI(a fyne.App, w fyne.Window) {
	ui := &mainUI{
		app:              a,
		win:              w,
		incomeCategories: append([]string{}, model.DefaultIncomeCategories...),
		expenseCategories: append([]string{}, model.DefaultExpenseCategories...),
		editingIndex:     -1,
	}

	ui.rebuildDisplay()

	// --- Selectores superiores ---
	currentMonth := months[time.Now().Month()-1]
	ui.monthSelect = widget.NewSelect(months, nil)
	ui.monthSelect.SetSelected(currentMonth)

	currentYear := time.Now().Year()
	var years []string
	for y := currentYear - 5; y <= currentYear + 5; y++ {
		years = append(years, strconv.Itoa(y))
	}

	ui.yearSelect = widget.NewSelect(years, nil)
	ui.yearSelect.SetSelected(strconv.Itoa(time.Now().Year()))

	ui.currencySelect = widget.NewSelect([]string{"USD", "CUP", "EUR"}, nil)
	ui.currencySelect.SetSelected("USD")

	btnReport := widget.NewButton("Generar Reporte", ui.onGenerateReport)
	btnReport.Importance = widget.HighImportance

	btnOpen := widget.NewButton("Abrir Excel", ui.onOpenExcel)
	btnClear := widget.NewButton("Limpiar", ui.onClear)

	topBar := container.NewHBox(
		widget.NewLabel("Mes:"), ui.monthSelect,
		widget.NewLabel("Año:"), ui.yearSelect,
		widget.NewLabel("Moneda:"), ui.currencySelect,
		btnReport,
		btnOpen,
		btnClear,
	)

	// --- Dos tablas separadas con pestañas ---
	ui.incomeTable = createEntryTable(ui, "Entrada")
	ui.expenseTable = createEntryTable(ui, "Gasto")

	tabs := container.NewAppTabs(
		container.NewTabItem("Entradas", ui.incomeTable),
		container.NewTabItem("Gastos", ui.expenseTable),
	)

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

	ui.invoiceEntry = widget.NewEntry()

	// Cargar historial completo desde preferencias
	ui.allDescriptions = a.Preferences().StringListWithFallback("descripciones", []string{})

	ui.descEntry = xwidget.NewCompletionEntry([]string{}) // empieza vacío

	// Filtrar sugerencias mientras se escribe
	ui.descEntry.OnChanged = func(s string) {
		query := strings.ToLower(strings.TrimSpace(s))
		if len(query) < 1 {
			ui.descEntry.HideCompletion()
			return
		}

		var filtered []string
		for _, d := range ui.allDescriptions {
			if strings.Contains(strings.ToLower(d), query) {
				filtered = append(filtered, d)
				if len(filtered) >= 10 { // limitar a 10 sugerencias
					break
				}
			}
		}

		if len(filtered) == 0 {
			ui.descEntry.HideCompletion()
			return
		}

		ui.descEntry.SetOptions(filtered)
		ui.descEntry.ShowCompletion()
	}

	ui.amountEntry = widget.NewEntry()
	ui.dayEntry = widget.NewEntry()

	form := widget.NewForm(
		widget.NewFormItem("Tipo", ui.typeRadio),
		widget.NewFormItem("Categoría", ui.categorySelect),
		widget.NewFormItem("# Factura", ui.invoiceEntry),
		widget.NewFormItem("Descripción", ui.descEntry),
		widget.NewFormItem("Monto", ui.amountEntry),
		widget.NewFormItem("Día del mes", ui.dayEntry),
	)

	ui.addButton = widget.NewButton("Añadir", ui.onAddOrUpdateEntry)
	ui.cancelButton = widget.NewButton("Cancelar", ui.onCancelEdit)
	ui.cancelButton.Hide()
	btnBox := container.NewHBox(ui.addButton, ui.cancelButton)

	formBox := container.NewVBox(
		widget.NewLabel("Añadir Entradas / Gastos :"),
		form,
		btnBox,
	)

	content := container.NewBorder(
		topBar,
		formBox,
		nil, nil,
		tabs,
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
	// Validar monto
	amountText := strings.ReplaceAll(strings.TrimSpace(ui.amountEntry.Text), ",", "")
	amount, err := strconv.ParseFloat(amountText, 64)
	if err != nil || amount < 0 {
		d := dialog.NewError(fmt.Errorf("Monto inválido"), ui.win)
		d.SetDismissText("Aceptar")
		d.Show()
		return
	}
	// Validar día
	day, err := strconv.Atoi(ui.dayEntry.Text)
	if err != nil || day < 1 || day > 31 {
		d := dialog.NewError(fmt.Errorf("Día inválido (1-31)"), ui.win)
		d.SetDismissText("Aceptar")
		d.Show()
		return
	}

	// Validar factura (opcional, pero si se ingresa debe ser entero positivo)
	invoiceText := strings.TrimSpace(ui.invoiceEntry.Text)
	var invoiceNum string

	if ui.editingIndex == -1 {
		if invoiceText == "" {
			invoiceNum = ""
		} else {
			invNum, err := strconv.Atoi(invoiceText)
			if err != nil || invNum <= 0 {
				d := dialog.NewError(fmt.Errorf("El número de factura debe ser un entero positivo (ej. 123)."), ui.win)
				d.SetDismissText("Aceptar")
				d.Show()
				return
			}
			invoiceNum = strconv.Itoa(invNum)
		}
	} else {
		if invoiceText == "" {
			invoiceNum = ui.entries[ui.editingIndex].InvoiceNum
		} else {
			invNum, err := strconv.Atoi(invoiceText)
			if err != nil || invNum <= 0 {
				d := dialog.NewError(fmt.Errorf("El número de factura debe ser un entero positivo (ej. 123)."), ui.win)
				d.SetDismissText("Aceptar")
				d.Show()
				return
			}
			invoiceNum = strconv.Itoa(invNum)
		}
	}

	year, _ := strconv.Atoi(ui.yearSelect.Selected)
	month := indexOfMonth(ui.monthSelect.Selected) + 1
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)

	entry := model.Entry{
		Date:        date,
		Type:        ui.typeRadio.Selected,
		Category:    ui.categorySelect.Selected,
		InvoiceNum:  invoiceNum,
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

	// Guardar descripción en el historial completo de autocompletado
	if entry.Description != "" {
		exists := false
		for _, d := range ui.allDescriptions {
			if d == entry.Description {
				exists = true
				break
			}
		}
		if !exists {
			ui.allDescriptions = append(ui.allDescriptions, entry.Description)
			ui.app.Preferences().SetStringList("descripciones", ui.allDescriptions)
		}
	}

	// Limpiar formulario
	ui.invoiceEntry.SetText("")
	ui.descEntry.SetText("")
	ui.amountEntry.SetText("")
	ui.dayEntry.SetText("1")
	ui.rebuildDisplay()
	ui.incomeTable.Refresh()
	ui.expenseTable.Refresh()
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
	ui.dayEntry.SetText(strconv.Itoa(entry.Date.Day()))
	ui.editingIndex = index
	ui.addButton.SetText("Actualizar")
	ui.cancelButton.Show()
}

func (ui *mainUI) deleteEntry(index int) {
    if index < 0 || index >= len(ui.entries) {
        return
    }
    d := dialog.NewConfirm("Eliminar", "¿Eliminar esta entrada?", func(ok bool) {
        if ok {
			ui.entries = append(ui.entries[:index], ui.entries[index+1:]...)
			ui.rebuildDisplay()
			ui.incomeTable.Refresh()
			ui.expenseTable.Refresh()
		}
    }, ui.win)
    d.SetConfirmText("Sí")
    d.SetDismissText("No")
    d.Show()
}

func (ui *mainUI) onCancelEdit() {
	ui.editingIndex = -1
	ui.addButton.SetText("Añadir")
	ui.cancelButton.Hide()
	ui.invoiceEntry.SetText("")
	ui.descEntry.SetText("")
	ui.amountEntry.SetText("")
	ui.dayEntry.SetText("1")
}

func (ui *mainUI) onGenerateReport() {
	if len(ui.entries) == 0 {
		dialog.ShowInformation("Sin datos", "No hay entradas para generar el reporte.", ui.win)
		return
	}
	month := ui.monthSelect.Selected
	year := ui.yearSelect.Selected
	currency := ui.currencySelect.Selected

	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			return
		}
		defer writer.Close()
		err = excel.GenerateReport(ui.entries, month, year, currency, ui.incomeCategories, ui.expenseCategories, writer)
		if err != nil {
			dialog.ShowError(err, ui.win)
		} else {
			dialog.ShowInformation("Éxito", "Reporte generado correctamente.", ui.win)
		}
	}, ui.win)

	fileName := fmt.Sprintf("reporte_%s_%s_%s.xlsx", month, year, currency)
	saveDialog.SetFileName(fileName)

	// Mostrar primero y redimensionar después
	saveDialog.Show()
	winSize := ui.win.Canvas().Size()
	saveDialog.Resize(fyne.NewSize(
		maxFloat32(winSize.Width*0.9, 900),
		maxFloat32(winSize.Height*0.85, 600),
	))
}

func (ui *mainUI) onOpenExcel() {
	d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()
		path := reader.URI().Path()
		entries, err := excel.ImportEntries(path)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Error al leer el archivo: %v", err), ui.win)
			return
		}
		ui.entries = entries

		// Añadir descripciones importadas al historial completo
		for _, e := range entries {
			if e.Description == "" {
				continue
			}
			exists := false
			for _, d := range ui.allDescriptions {
				if d == e.Description {
					exists = true
					break
				}
			}
			if !exists {
				ui.allDescriptions = append(ui.allDescriptions, e.Description)
			}
		}
		ui.app.Preferences().SetStringList("descripciones", ui.allDescriptions)

		ui.rebuildDisplay()
		ui.incomeTable.Refresh()
		ui.expenseTable.Refresh()

		if len(entries) > 0 {
			firstCurrency := entries[0].Currency
			if firstCurrency == "USD" || firstCurrency == "EUR" || firstCurrency == "CUP" {
				ui.currencySelect.SetSelected(firstCurrency)
			}
		}

		if len(entries) > 0 {
			firstDate := entries[0].Date
			ui.monthSelect.SetSelected(months[firstDate.Month()-1])
			ui.yearSelect.SetSelected(strconv.Itoa(firstDate.Year()))
		}

		dialog.ShowInformation("Cargado", fmt.Sprintf("Se cargaron %d entradas correctamente.", len(entries)), ui.win)
	}, ui.win)

	// Mostrar primero y redimensionar después
	d.Show()
	winSize := ui.win.Canvas().Size()
	d.Resize(fyne.NewSize(
		maxFloat32(winSize.Width*0.9, 900),
		maxFloat32(winSize.Height*0.85, 600),
	))
}

func (ui *mainUI) onClear() {
	if len(ui.entries) == 0 {
		return // nada que limpiar
	}
	d := dialog.NewConfirm("Limpiar todo", "¿Estás seguro de limpiar todas las entradas?", func(ok bool) {
		if ok {
			ui.entries = nil
			ui.rebuildDisplay()
			ui.incomeTable.Refresh()
			ui.expenseTable.Refresh()
			// Restablecer selectores a valores por defecto
			currentMonth := months[time.Now().Month()-1]
			ui.monthSelect.SetSelected(currentMonth)
			ui.yearSelect.SetSelected(strconv.Itoa(time.Now().Year()))
			ui.currencySelect.SetSelected("USD")
		}
	}, ui.win)
	d.SetConfirmText("Sí")
	d.SetDismissText("No")
	d.Show()
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

// maxFloat32 devuelve el mayor de dos float32.
func maxFloat32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// formatAmountDisplay formatea un monto como "24,250.25 USD"
// (comas para miles, punto para decimales).
func formatAmountDisplay(amount float64, currency string) string {
	str := strconv.FormatFloat(amount, 'f', 2, 64) // ej. "24250.25"
	neg := false
	if strings.HasPrefix(str, "-") {
		neg = true
		str = str[1:]
	}
	parts := strings.Split(str, ".")
	intPart, decPart := parts[0], parts[1]

	// Insertar comas cada 3 dígitos en la parte entera
	if len(intPart) > 3 {
		for i := len(intPart) - 3; i > 0; i -= 3 {
			intPart = intPart[:i] + "," + intPart[i:]
		}
	}

	res := intPart + "." + decPart + " " + currency
	if neg {
		res = "-" + res
	}
	return res
}

func sortIndexedByCategory(entries []indexedEntry, order []string, byInvoice bool) {
	rank := make(map[string]int)
	for i, cat := range order {
		rank[cat] = i
	}
	sort.Slice(entries, func(i, j int) bool {
		ci, cj := entries[i].entry.Category, entries[j].entry.Category
		ri, oki := rank[ci]
		rj, okj := rank[cj]
		if oki && okj {
			if ri != rj {
				return ri < rj
			}
		} else if oki {
			return true
		} else if okj {
			return false
		}
		if ci != cj {
			return ci < cj
		}
		// Misma categoría
		if byInvoice {
			return lessByInvoice(entries[i].entry, entries[j].entry)
		}
		return entries[i].entry.Date.Before(entries[j].entry.Date)
	})
}

// lessByInvoice compara dos entradas por número de factura (numérico).
// Las entradas sin factura válida van al final; si empatan, por fecha.
func lessByInvoice(a, b model.Entry) bool {
	ai, aerr := strconv.Atoi(strings.TrimSpace(a.InvoiceNum))
	bi, berr := strconv.Atoi(strings.TrimSpace(b.InvoiceNum))

	if aerr == nil && berr == nil {
		if ai != bi {
			return ai < bi
		}
		return a.Date.Before(b.Date)
	}
	if aerr == nil && berr != nil {
		return true // a tiene factura, b no → a primero
	}
	if aerr != nil && berr == nil {
		return false
	}
	// Ninguno tiene factura válida → por fecha
	return a.Date.Before(b.Date)
}

func (ui *mainUI) rebuildDisplay() {

    var all []indexedEntry
    for i, e := range ui.entries {
        all = append(all, indexedEntry{e, i})
    }

    var income, expense []indexedEntry
    for _, ie := range all {
        if ie.entry.Type == "Entrada" {
            income = append(income, ie)
        } else {
            expense = append(expense, ie)
        }
    }

    sortIndexedByCategory(income, ui.incomeCategories, true)  // entradas → por factura
	sortIndexedByCategory(expense, ui.expenseCategories, false) // gastos → por fecha

    // Llenar slices combinados (por si aún se usan)
    ui.displayEntries = ui.displayEntries[:0]
    ui.displayOriginalIndices = ui.displayOriginalIndices[:0]
    ui.displayIncomeEntries = ui.displayIncomeEntries[:0]
    ui.displayIncomeIndices = ui.displayIncomeIndices[:0]
    ui.displayExpenseEntries = ui.displayExpenseEntries[:0]
    ui.displayExpenseIndices = ui.displayExpenseIndices[:0]

    for _, ie := range income {
        ui.displayIncomeEntries = append(ui.displayIncomeEntries, ie.entry)
        ui.displayIncomeIndices = append(ui.displayIncomeIndices, ie.idx)
        ui.displayEntries = append(ui.displayEntries, ie.entry)
        ui.displayOriginalIndices = append(ui.displayOriginalIndices, ie.idx)
    }
    for _, ie := range expense {
        ui.displayExpenseEntries = append(ui.displayExpenseEntries, ie.entry)
        ui.displayExpenseIndices = append(ui.displayExpenseIndices, ie.idx)
        ui.displayEntries = append(ui.displayEntries, ie.entry)
        ui.displayOriginalIndices = append(ui.displayOriginalIndices, ie.idx)
    }
}

func createEntryTable(ui *mainUI, tipo string) *widget.Table {
    table := widget.NewTable(
        func() (int, int) {
            if tipo == "Entrada" {
                return len(ui.displayIncomeEntries) + 1, 9
            }
            return len(ui.displayExpenseEntries) + 1, 9
        },
        func() fyne.CanvasObject {
            bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
            label := widget.NewLabel("")
            label.Truncation = fyne.TextTruncateEllipsis
            return container.NewStack(bg, label)
        },
        func(id widget.TableCellID, obj fyne.CanvasObject) {
            stack := obj.(*fyne.Container)
            bg := stack.Objects[0].(*canvas.Rectangle)

            if len(stack.Objects) > 1 {
                stack.Remove(stack.Objects[1])
            }

            var content fyne.CanvasObject

            if id.Row == 0 {
                headers := []string{"Nº", "Fecha", "Tipo", "Categoría", "# Factura", "Descripción", "Monto", "Editar", "Eliminar"}
                label := widget.NewLabel(headers[id.Col])
                label.TextStyle = fyne.TextStyle{Bold: true}
                label.Truncation = fyne.TextTruncateEllipsis
                content = label
                bg.FillColor = theme.Color(theme.ColorNameHeaderBackground)
            } else {
                var entry model.Entry
                var originalIndex int
                if tipo == "Entrada" {
                    entry = ui.displayIncomeEntries[id.Row-1]
                    originalIndex = ui.displayIncomeIndices[id.Row-1]
                } else {
                    entry = ui.displayExpenseEntries[id.Row-1]
                    originalIndex = ui.displayExpenseIndices[id.Row-1]
                }

                if id.Col < 7 {
                    label := widget.NewLabel("")
                    label.Truncation = fyne.TextTruncateEllipsis
                    switch id.Col {
                    case 0:
                        label.SetText(strconv.Itoa(id.Row))
                    case 1:
                        label.SetText(entry.Date.Format("02/01/2006"))
                    case 2:
                        label.SetText(entry.Type)
                    case 3:
                        label.SetText(entry.Category)
                    case 4:
                        label.SetText(entry.InvoiceNum)
                    case 5:
                        label.SetText(entry.Description)
                    case 6:
                        label.SetText(formatAmountDisplay(entry.Amount, entry.Currency))
                    }
                    content = label
                } else if id.Col == 7 {
                    btn := widget.NewButton("Editar", func() {
                        ui.startEdit(originalIndex)
                    })
                    btn.Importance = widget.LowImportance
                    content = btn
                } else if id.Col == 8 {
                    btn := widget.NewButton("Eliminar", func() {
                        ui.deleteEntry(originalIndex)
                    })
                    btn.Importance = widget.DangerImportance
                    content = btn
                }

                // Colores según tipo
                if id.Row%2 == 0 {
                    if entry.Type == "Entrada" {
                        bg.FillColor = color.NRGBA{R: 198, G: 239, B: 206, A: 255}
                    } else {
                        bg.FillColor = color.NRGBA{R: 255, G: 199, B: 206, A: 255}
                    }
                } else {
                    if entry.Type == "Entrada" {
                        bg.FillColor = color.NRGBA{R: 217, G: 242, B: 217, A: 255}
                    } else {
                        bg.FillColor = color.NRGBA{R: 255, G: 205, B: 210, A: 255}
                    }
                }
            }

            bg.Refresh()
            stack.Add(content)
        },
    )

    // Anchos
    table.SetColumnWidth(0, 100)
    table.SetColumnWidth(1, 150)
    table.SetColumnWidth(2, 100)
    table.SetColumnWidth(3, 150)
    table.SetColumnWidth(4, 150)
    table.SetColumnWidth(5, 750)
    table.SetColumnWidth(6, 250)
    table.SetColumnWidth(7, 100)
    table.SetColumnWidth(8, 100)

    return table
}