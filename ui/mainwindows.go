package ui

import (
	"fmt"
	"strconv"
	"strings"
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

// Meses del año (global para poder usarla en onClear)
var months = []string{
	"Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio",
	"Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre",
}

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
	dayEntry          *widget.Entry
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
	currentMonth := months[time.Now().Month()-1]
	ui.monthSelect = widget.NewSelect(months, nil)
	ui.monthSelect.SetSelected(currentMonth)

	// Generar lista de años: desde hace 5 años hasta dentro de 5 años
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

	// --- Tabla con botones reales para editar/eliminar y texto truncado ---
	ui.table = widget.NewTable(
		func() (int, int) { return len(ui.entries) + 1, 9 },
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
			// Label base con truncamiento habilitado
			label := widget.NewLabel("")
			label.Truncation = fyne.TextTruncateEllipsis
			return container.NewStack(bg, label)
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			stack := obj.(*fyne.Container)
			bg := stack.Objects[0].(*canvas.Rectangle)

			// Reemplazar contenido preservando el fondo
			if len(stack.Objects) > 1 {
				stack.Remove(stack.Objects[1])
			}

			var content fyne.CanvasObject

			if id.Row == 0 {
				// Cabecera
				headers := []string{"Nº", "Fecha", "Tipo", "Categoría", "# Factura", "Descripción", "Monto", "Editar", "Eliminar"}
				label := widget.NewLabel(headers[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
				label.Truncation = fyne.TextTruncateEllipsis
				content = label
				bg.FillColor = theme.Color(theme.ColorNameHeaderBackground)
			} else {
				entry := ui.entries[id.Row-1]

				if id.Col < 7 {
					// Columnas de datos con truncamiento
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
						ui.startEdit(id.Row - 1)
					})
					btn.Importance = widget.LowImportance
					content = btn
				} else if id.Col == 8 {
					btn := widget.NewButton("Eliminar", func() {
						ui.deleteEntry(id.Row - 1)
					})
					btn.Importance = widget.DangerImportance
					content = btn
				}

				// Color de fondo alternado
				if id.Row%2 == 0 {
					bg.FillColor = theme.Color(theme.ColorNameHover)
				} else {
					bg.FillColor = theme.Color(theme.ColorNameBackground)
				}
			}

			bg.Refresh()
			stack.Add(content)
		},
	)

	// Anchos de columna ajustados
	ui.table.SetColumnWidth(0, 100)  // Nº
	ui.table.SetColumnWidth(1, 150) // Fecha
	ui.table.SetColumnWidth(2, 100)  // Tipo
	ui.table.SetColumnWidth(3, 150) // Categoría
	ui.table.SetColumnWidth(4, 150) // # Factura
	ui.table.SetColumnWidth(5, 750) // Descripción (más ancha)
	ui.table.SetColumnWidth(6, 250) // Monto
	ui.table.SetColumnWidth(7, 100)  // Editar
	ui.table.SetColumnWidth(8, 100)  // Eliminar

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
	ui.invoiceEntry.SetPlaceHolder("# Factura")
	ui.descEntry = widget.NewEntry()
	ui.descEntry.SetPlaceHolder("Descripción")
	ui.amountEntry = widget.NewEntry()
	ui.amountEntry.SetPlaceHolder("Monto (ej. 150.00)")
	ui.dayEntry = widget.NewEntry()
	ui.dayEntry.SetPlaceHolder("Día (1-31)")
	ui.dayEntry.SetText("1")

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
		widget.NewLabel("Nueva entrada:"),
		form,
		btnBox,
	)

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
    // Validar monto
    amount, err := strconv.ParseFloat(ui.amountEntry.Text, 64)
    if err != nil || amount <= 0 {
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
		// Nueva entrada: factura opcional
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
		// Edición: si vacío, conservar el anterior; si no, validar
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

    // Limpiar formulario
    ui.invoiceEntry.SetText("")
    ui.descEntry.SetText("")
    ui.amountEntry.SetText("")
    ui.dayEntry.SetText("1")
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
            ui.table.Refresh()
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
		err = excel.GenerateReport(ui.entries, month, year, currency, writer)
		if err != nil {
			dialog.ShowError(err, ui.win)
		} else {
			dialog.ShowInformation("Éxito", "Reporte generado correctamente.", ui.win)
		}
	}, ui.win)
	fileName := fmt.Sprintf("reporte_%s_%s_%s.xlsx", month, year, currency)
	saveDialog.SetFileName(fileName)
	saveDialog.Show()
}

func (ui *mainUI) onOpenExcel() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
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

		// Si se importaron datos, ajustar el selector de moneda al del primer registro
		if len(entries) > 0 {
			firstCurrency := entries[0].Currency
			// Asegurarse de que el valor es válido (USD, EUR, CUP)
			if firstCurrency == "USD" || firstCurrency == "EUR" || firstCurrency == "CUP" {
				ui.currencySelect.SetSelected(firstCurrency)
			}
		}

		// sincronizar mes y año con la fecha del primer registro
		if len(entries) > 0 {
			firstDate := entries[0].Date
			months := []string{"Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio",
				"Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre"}
			ui.monthSelect.SetSelected(months[firstDate.Month()-1])
			ui.yearSelect.SetSelected(strconv.Itoa(firstDate.Year()))
		}
		ui.table.Refresh()
		dialog.ShowInformation("Cargado", fmt.Sprintf("Se cargaron %d entradas correctamente.", len(entries)), ui.win)
	}, ui.win)
}

func (ui *mainUI) onClear() {
	if len(ui.entries) == 0 {
		return // nada que limpiar
	}
	d := dialog.NewConfirm("Limpiar todo", "¿Estás seguro de eliminar todas las entradas?", func(ok bool) {
		if ok {
			ui.entries = nil
			ui.table.Refresh()
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

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// formatAmountDisplay formatea un monto como "200,000 USD" (coma para miles, sin decimales)
func formatAmountDisplay(amount float64, currency string) string {
	// Redondear a entero si no hay decimales significativos, pero normalmente trabajaremos con enteros
	intAmount := int64(amount)
	str := strconv.FormatInt(intAmount, 10)
	// Insertar comas cada 3 dígitos
	if len(str) > 3 {
		for i := len(str) - 3; i > 0; i -= 3 {
			str = str[:i] + "," + str[i:]
		}
	}
	return str + " " + currency
}