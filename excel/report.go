package excel

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"marmur_excel_tool/model"

	"github.com/xuri/excelize/v2"
)

func GenerateReport(entries []model.Entry, month, year, currency string, w io.Writer) error {
	f := excelize.NewFile()
	
	// Filtrar por moneda
	var filtered []model.Entry
	for _, e := range entries {
		if e.Currency == currency {
			filtered = append(filtered, e)
		}
	}
	if len(filtered) == 0 {
		return fmt.Errorf("no hay datos para la moneda %s", currency)
	}

	// ===== ESTILOS =====
	
	// Estilo título principal
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   14,
			Color:  "1F4E79",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"D6E4F0"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "bottom", Color: "1F4E79", Style: 2},
		},
	})
	
	// Estilo cabecera (azul oscuro)
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   11,
			Color:  "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"2E75B6"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "FFFFFF", Style: 1},
			{Type: "right", Color: "FFFFFF", Style: 1},
			{Type: "top", Color: "FFFFFF", Style: 1},
			{Type: "bottom", Color: "FFFFFF", Style: 1},
		},
	})
	
	// Estilo resumen (verde claro para entradas)
	incomeStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  11,
			Color: "006100",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"C6EFCE"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "006100", Style: 1},
			{Type: "right", Color: "006100", Style: 1},
			{Type: "top", Color: "006100", Style: 1},
			{Type: "bottom", Color: "006100", Style: 1},
		},
	})
	
	// Estilo resumen (rojo claro para gastos)
	expenseStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  11,
			Color: "9C0006",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"FFC7CE"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "9C0006", Style: 1},
			{Type: "right", Color: "9C0006", Style: 1},
			{Type: "top", Color: "9C0006", Style: 1},
			{Type: "bottom", Color: "9C0006", Style: 1},
		},
	})
	
	// Estilo balance (azul claro)
	balanceStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   12,
			Color:  "1F4E79",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"BDD7EE"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "1F4E79", Style: 2},
			{Type: "right", Color: "1F4E79", Style: 2},
			{Type: "top", Color: "1F4E79", Style: 2},
			{Type: "bottom", Color: "1F4E79", Style: 2},
		},
	})
	
	// Estilo datos normales con bordes
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "B4C6E7", Style: 1},
			{Type: "right", Color: "B4C6E7", Style: 1},
			{Type: "top", Color: "B4C6E7", Style: 1},
			{Type: "bottom", Color: "B4C6E7", Style: 1},
		},
	})
	
	// Estilo filas alternadas (gris claro)
	altRowStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"F2F2F2"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "B4C6E7", Style: 1},
			{Type: "right", Color: "B4C6E7", Style: 1},
			{Type: "top", Color: "B4C6E7", Style: 1},
			{Type: "bottom", Color: "B4C6E7", Style: 1},
		},
	})
	
	// Estilo subtítulo sección
	subtitleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  11,
			Color: "2E75B6",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"E8F0FE"},
			Pattern: 1,
		},
	})

	// ===== HOJA 1: RESUMEN GENERAL =====
	f.SetSheetName("Sheet1", "Resumen")
	
	// Título
	f.SetCellValue("Resumen", "A1", fmt.Sprintf("REPORTE MENSUAL - %s %s", month, year))
	f.MergeCell("Resumen", "A1", "C1")
	f.SetCellStyle("Resumen", "A1", "C1", titleStyle)
	
	f.SetCellValue("Resumen", "A2", fmt.Sprintf("Moneda: %s", currency))
	f.MergeCell("Resumen", "A2", "C2")
	f.SetCellStyle("Resumen", "A2", "C2", subtitleStyle)
	
	// Cabeceras
	f.SetCellValue("Resumen", "A4", "Concepto")
	f.SetCellValue("Resumen", "B4", "Tipo")
	f.SetCellValue("Resumen", "C4", "Monto")
	f.SetCellStyle("Resumen", "A4", "C4", headerStyle)
	
	// Totales
	var totalIncome, totalExpense float64
	for _, e := range filtered {
		if e.Type == "Entrada" {
			totalIncome += e.Amount
		} else {
			totalExpense += e.Amount
		}
	}
	balance := totalIncome - totalExpense
	
	// Filas de resumen con colores
	f.SetCellValue("Resumen", "A5", "Total Entradas")
	f.SetCellValue("Resumen", "B5", "Ingreso")
	f.SetCellValue("Resumen", "C5", totalIncome)
	f.SetCellStyle("Resumen", "A5", "C5", incomeStyle)
	f.SetCellValue("Resumen", "C5", formatCurrency(totalIncome, currency))
	
	f.SetCellValue("Resumen", "A6", "Total Gastos")
	f.SetCellValue("Resumen", "B6", "Egreso")
	f.SetCellValue("Resumen", "C6", totalExpense)
	f.SetCellStyle("Resumen", "A6", "C6", expenseStyle)
	f.SetCellValue("Resumen", "C6", formatCurrency(totalExpense, currency))
	
	// Balance con color según sea positivo o negativo
	f.SetCellValue("Resumen", "A8", "BALANCE FINAL")
	f.MergeCell("Resumen", "A8", "C8")
	f.SetCellStyle("Resumen", "A8", "C8", balanceStyle)
	
	f.SetCellValue("Resumen", "A9", "Resultado:")
	f.SetCellValue("Resumen", "C9", formatCurrency(balance, currency))
	if balance >= 0 {
		f.SetCellStyle("Resumen", "A9", "C9", incomeStyle)
	} else {
		f.SetCellStyle("Resumen", "A9", "C9", expenseStyle)
	}
	
	// Ancho columnas
	f.SetColWidth("Resumen", "A", "A", 25)
	f.SetColWidth("Resumen", "B", "B", 15)
	f.SetColWidth("Resumen", "C", "C", 20)
	
	// ===== HOJA 2: POR CATEGORÍAS =====
	f.NewSheet("Por Categorías")
	
	// Título
	f.SetCellValue("Por Categorías", "A1", "RESUMEN POR CATEGORÍAS")
	f.MergeCell("Por Categorías", "A1", "C1")
	f.SetCellStyle("Por Categorías", "A1", "C1", titleStyle)
	
	// Cabeceras
	f.SetCellValue("Por Categorías", "A3", "Categoría")
	f.SetCellValue("Por Categorías", "B3", "Tipo")
	f.SetCellValue("Por Categorías", "C3", "Monto")
	f.SetCellStyle("Por Categorías", "A3", "C3", headerStyle)
	
	// Calcular por categoría
	incomeByCat := make(map[string]float64)
	expenseByCat := make(map[string]float64)
	for _, e := range filtered {
		if e.Type == "Entrada" {
			incomeByCat[e.Category] += e.Amount
		} else {
			expenseByCat[e.Category] += e.Amount
		}
	}
	
	row := 4
	// Entradas primero
	if len(incomeByCat) > 0 {
		f.SetCellValue("Por Categorías", fmt.Sprintf("A%d", row), "ENTRADAS")
		f.MergeCell("Por Categorías", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row))
		f.SetCellStyle("Por Categorías", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), subtitleStyle)
		row++
		
		for cat, val := range incomeByCat {
			f.SetCellValue("Por Categorías", fmt.Sprintf("A%d", row), cat)
			f.SetCellValue("Por Categorías", fmt.Sprintf("B%d", row), "Entrada")
			f.SetCellValue("Por Categorías", fmt.Sprintf("C%d", row), formatCurrency(val, currency))
			if (row-4)%2 == 0 {
				f.SetCellStyle("Por Categorías", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), altRowStyle)
			} else {
				f.SetCellStyle("Por Categorías", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), dataStyle)
			}
			row++
		}
		
		// Total entradas
		f.SetCellValue("Por Categorías", fmt.Sprintf("A%d", row), "Total Entradas")
		f.SetCellValue("Por Categorías", fmt.Sprintf("C%d", row), formatCurrency(totalIncome, currency))
		f.SetCellStyle("Por Categorías", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), incomeStyle)
		row += 2
	}
	
	// Luego gastos
	if len(expenseByCat) > 0 {
		f.SetCellValue("Por Categorías", fmt.Sprintf("A%d", row), "GASTOS")
		f.MergeCell("Por Categorías", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row))
		f.SetCellStyle("Por Categorías", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), subtitleStyle)
		row++
		
		for cat, val := range expenseByCat {
			f.SetCellValue("Por Categorías", fmt.Sprintf("A%d", row), cat)
			f.SetCellValue("Por Categorías", fmt.Sprintf("B%d", row), "Gasto")
			f.SetCellValue("Por Categorías", fmt.Sprintf("C%d", row), formatCurrency(val, currency))
			if (row-4)%2 == 0 {
				f.SetCellStyle("Por Categorías", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), altRowStyle)
			} else {
				f.SetCellStyle("Por Categorías", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), dataStyle)
			}
			row++
		}
		
		// Total gastos
		f.SetCellValue("Por Categorías", fmt.Sprintf("A%d", row), "Total Gastos")
		f.SetCellValue("Por Categorías", fmt.Sprintf("C%d", row), formatCurrency(totalExpense, currency))
		f.SetCellStyle("Por Categorías", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), expenseStyle)
	}
	
	// Balance final en esta hoja también
	row += 2
	f.SetCellValue("Por Categorías", fmt.Sprintf("A%d", row), "BALANCE")
	f.SetCellValue("Por Categorías", fmt.Sprintf("C%d", row), formatCurrency(balance, currency))
	if balance >= 0 {
		f.SetCellStyle("Por Categorías", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), incomeStyle)
	} else {
		f.SetCellStyle("Por Categorías", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), expenseStyle)
	}
	
	f.SetColWidth("Por Categorías", "A", "A", 25)
	f.SetColWidth("Por Categorías", "B", "B", 15)
	f.SetColWidth("Por Categorías", "C", "C", 20)
	
	// ===== HOJA 3: LISTADO DETALLADO =====
	f.NewSheet("Listado")
	
	// Título
	f.SetCellValue("Listado", "A1", "LISTADO DETALLADO DE TRANSACCIONES")
	f.MergeCell("Listado", "A1", "F1")
	f.SetCellStyle("Listado", "A1", "F1", titleStyle)
	
	// Cabeceras
	headers := []string{"Fecha", "Tipo", "Categoría", "# Factura", "Descripción", "Monto"}
	for i, h := range headers {
		cell := fmt.Sprintf("%c3", 'A'+i)
		f.SetCellValue("Listado", cell, h)
	}
	f.SetCellStyle("Listado", "A3", "F3", headerStyle)
	
	// Datos con filas alternadas
	for i, e := range filtered {
		row := i + 4
		f.SetCellValue("Listado", fmt.Sprintf("A%d", row), e.Date.Format("02/01/2006"))
		f.SetCellValue("Listado", fmt.Sprintf("B%d", row), e.Type)
		f.SetCellValue("Listado", fmt.Sprintf("C%d", row), e.Category)
		f.SetCellValue("Listado", fmt.Sprintf("D%d", row), e.InvoiceNum)
		f.SetCellValue("Listado", fmt.Sprintf("E%d", row), e.Description)
		f.SetCellValue("Listado", fmt.Sprintf("F%d", row), formatCurrency(e.Amount, currency))
		
		// Color según tipo y fila alternada
		if e.Type == "Entrada" {
			if i%2 == 0 {
				// Estilo combinado para entradas filas pares
				entryAltStyle, _ := f.NewStyle(&excelize.Style{
					Fill: excelize.Fill{
						Type:    "pattern",
						Color:   []string{"E8F5E9"},
						Pattern: 1,
					},
					Border: []excelize.Border{
						{Type: "left", Color: "C8E6C9", Style: 1},
						{Type: "right", Color: "C8E6C9", Style: 1},
						{Type: "top", Color: "C8E6C9", Style: 1},
						{Type: "bottom", Color: "C8E6C9", Style: 1},
					},
				})
				f.SetCellStyle("Listado", fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row), entryAltStyle)
			}
		} else {
			if i%2 == 0 {
				// Estilo combinado para gastos filas pares
				expenseAltStyle, _ := f.NewStyle(&excelize.Style{
					Fill: excelize.Fill{
						Type:    "pattern",
						Color:   []string{"FFEBEE"},
						Pattern: 1,
					},
					Border: []excelize.Border{
						{Type: "left", Color: "FFCDD2", Style: 1},
						{Type: "right", Color: "FFCDD2", Style: 1},
						{Type: "top", Color: "FFCDD2", Style: 1},
						{Type: "bottom", Color: "FFCDD2", Style: 1},
					},
				})
				f.SetCellStyle("Listado", fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row), expenseAltStyle)
			}
		}
	}
	
	f.SetColWidth("Listado", "A", "A", 12)
	f.SetColWidth("Listado", "B", "B", 10)
	f.SetColWidth("Listado", "C", "C", 15)
	f.SetColWidth("Listado", "D", "D", 12)
	f.SetColWidth("Listado", "E", "E", 40)
	f.SetColWidth("Listado", "F", "F", 15)
	
	// Activar primera hoja
	f.SetActiveSheet(0)
	
	return f.Write(w)
}

// ImportEntries lee un archivo Excel con el formato de guardado
func ImportEntries(path string) ([]model.Entry, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir el archivo: %v", err)
	}
	defer f.Close()

	// Intentar leer la hoja "Listado"
	rows, err := f.GetRows("Listado")
	if err != nil {
		// Si no existe "Listado", intentar con la primera hoja
		sheets := f.GetSheetList()
		if len(sheets) > 0 {
			rows, err = f.GetRows(sheets[0])
			if err != nil {
				return nil, fmt.Errorf("no se pudieron leer los datos: %v", err)
			}
		} else {
			return nil, fmt.Errorf("el archivo no contiene hojas")
		}
	}
	
	var entries []model.Entry
	for i, row := range rows {
		if i == 0 {
			// Saltar cabecera
			continue
		}
		if len(row) < 6 {
			continue
		}
		
		date, err := time.Parse("02/01/2006", row[0])
		if err != nil {
			// Intentar otros formatos de fecha
			date, err = time.Parse("2006-01-02", row[0])
			if err != nil {
				continue // Saltar fila si la fecha no es válida
			}
		}
		
		amount, err := strconv.ParseFloat(row[5], 64)
		if err != nil {
			amount = 0
		}
		
		currency := "USD" // Por defecto
		if len(row) > 6 {
			currency = row[6]
		}
		
		entries = append(entries, model.Entry{
			Date:        date,
			Type:        row[1],
			Category:    row[2],
			InvoiceNum:  row[3],
			Description: row[4],
			Amount:      amount,
			Currency:    currency,
		})
	}
	
	if len(entries) == 0 {
		return nil, fmt.Errorf("no se encontraron entradas válidas en el archivo")
	}
	
	return entries, nil
}
// SaveEntries guarda las entradas en un Excel intermedio
func SaveEntries(entries []model.Entry, w io.Writer) error {
	f := excelize.NewFile()
	f.SetSheetName("Sheet1", "Listado")
	
	// Cabeceras con estilo
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"4472C4"},
			Pattern: 1,
		},
	})
	
	headers := []string{"Fecha", "Tipo", "Categoría", "# Factura", "Descripción", "Monto", "Moneda"}
	for i, h := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue("Listado", cell, h)
	}
	f.SetCellStyle("Listado", "A1", "G1", headerStyle)
	
	// Datos
	for i, e := range entries {
		row := i + 2
		f.SetCellValue("Listado", fmt.Sprintf("A%d", row), e.Date.Format("02/01/2006"))
		f.SetCellValue("Listado", fmt.Sprintf("B%d", row), e.Type)
		f.SetCellValue("Listado", fmt.Sprintf("C%d", row), e.Category)
		f.SetCellValue("Listado", fmt.Sprintf("D%d", row), e.InvoiceNum)
		f.SetCellValue("Listado", fmt.Sprintf("E%d", row), e.Description)
		f.SetCellValue("Listado", fmt.Sprintf("F%d", row), e.Amount)
		f.SetCellValue("Listado", fmt.Sprintf("G%d", row), e.Currency)
	}
	
	f.SetColWidth("Listado", "A", "A", 12)
	f.SetColWidth("Listado", "B", "B", 10)
	f.SetColWidth("Listado", "C", "C", 15)
	f.SetColWidth("Listado", "D", "D", 12)
	f.SetColWidth("Listado", "E", "E", 40)
	f.SetColWidth("Listado", "F", "F", 12)
	f.SetColWidth("Listado", "G", "G", 10)
	
	return f.Write(w)
}

// Formatea montos con símbolo de moneda
func formatCurrency(amount float64, currency string) string {
	symbol := "$"
	switch currency {
	case "EUR":
		symbol = "€"
	case "CUP":
		symbol = "CUP "
	}
	return fmt.Sprintf("%s %.2f", symbol, amount)
}