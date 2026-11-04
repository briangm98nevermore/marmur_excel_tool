package excel

import (
	"fmt"
	"io"
	"marmur_excel_tool/model"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

func GenerateReport(entries []model.Entry, month, year, currency string,
    incomeOrder, expenseOrder []string, w io.Writer) error {

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

    // ===== TODOS LOS ESTILOS =====
    titleStyle, _ := f.NewStyle(&excelize.Style{
        Font:      &excelize.Font{Bold: true, Size: 14, Color: "1F4E79"},
        Fill:      excelize.Fill{Type: "pattern", Color: []string{"D6E4F0"}, Pattern: 1},
        Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
        Border:    []excelize.Border{{Type: "bottom", Color: "1F4E79", Style: 2}},
    })

    headerStyle, _ := f.NewStyle(&excelize.Style{
        Font:      &excelize.Font{Bold: true, Size: 11, Color: "FFFFFF"},
        Fill:      excelize.Fill{Type: "pattern", Color: []string{"2E75B6"}, Pattern: 1},
        Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
        Border: []excelize.Border{
            {Type: "left", Color: "FFFFFF", Style: 1},
            {Type: "right", Color: "FFFFFF", Style: 1},
            {Type: "top", Color: "FFFFFF", Style: 1},
            {Type: "bottom", Color: "FFFFFF", Style: 1},
        },
    })

    incomeStyle, _ := f.NewStyle(&excelize.Style{
        Font:   &excelize.Font{Bold: true, Size: 11, Color: "006100"},
        Fill:   excelize.Fill{Type: "pattern", Color: []string{"C6EFCE"}, Pattern: 1},
        Border: []excelize.Border{
            {Type: "left", Color: "006100", Style: 1},
            {Type: "right", Color: "006100", Style: 1},
            {Type: "top", Color: "006100", Style: 1},
            {Type: "bottom", Color: "006100", Style: 1},
        },
    })

    expenseStyle, _ := f.NewStyle(&excelize.Style{
        Font:   &excelize.Font{Bold: true, Size: 11, Color: "9C0006"},
        Fill:   excelize.Fill{Type: "pattern", Color: []string{"FFC7CE"}, Pattern: 1},
        Border: []excelize.Border{
            {Type: "left", Color: "9C0006", Style: 1},
            {Type: "right", Color: "9C0006", Style: 1},
            {Type: "top", Color: "9C0006", Style: 1},
            {Type: "bottom", Color: "9C0006", Style: 1},
        },
    })

    balanceStyle, _ := f.NewStyle(&excelize.Style{
        Font:      &excelize.Font{Bold: true, Size: 12, Color: "1F4E79"},
        Fill:      excelize.Fill{Type: "pattern", Color: []string{"BDD7EE"}, Pattern: 1},
        Alignment: &excelize.Alignment{Horizontal: "center"},
        Border: []excelize.Border{
            {Type: "left", Color: "1F4E79", Style: 2},
            {Type: "right", Color: "1F4E79", Style: 2},
            {Type: "top", Color: "1F4E79", Style: 2},
            {Type: "bottom", Color: "1F4E79", Style: 2},
        },
    })

    dataStyle, _ := f.NewStyle(&excelize.Style{
        Border: []excelize.Border{
            {Type: "left", Color: "B4C6E7", Style: 1},
            {Type: "right", Color: "B4C6E7", Style: 1},
            {Type: "top", Color: "B4C6E7", Style: 1},
            {Type: "bottom", Color: "B4C6E7", Style: 1},
        },
    })

    altRowStyle, _ := f.NewStyle(&excelize.Style{
        Fill: excelize.Fill{Type: "pattern", Color: []string{"F2F2F2"}, Pattern: 1},
        Border: []excelize.Border{
            {Type: "left", Color: "B4C6E7", Style: 1},
            {Type: "right", Color: "B4C6E7", Style: 1},
            {Type: "top", Color: "B4C6E7", Style: 1},
            {Type: "bottom", Color: "B4C6E7", Style: 1},
        },
    })

    subtitleStyle, _ := f.NewStyle(&excelize.Style{
        Font: &excelize.Font{Bold: true, Size: 11, Color: "2E75B6"},
        Fill: excelize.Fill{Type: "pattern", Color: []string{"E8F0FE"}, Pattern: 1},
    })

    incomeRowEven, _ := f.NewStyle(&excelize.Style{
        Fill: excelize.Fill{Type: "pattern", Color: []string{"C6EFCE"}, Pattern: 1},
        Border: []excelize.Border{
            {Type: "left", Color: "A5D6A7", Style: 1},
            {Type: "right", Color: "A5D6A7", Style: 1},
            {Type: "top", Color: "A5D6A7", Style: 1},
            {Type: "bottom", Color: "A5D6A7", Style: 1},
        },
    })
    incomeRowOdd, _ := f.NewStyle(&excelize.Style{
        Fill: excelize.Fill{Type: "pattern", Color: []string{"D9F2D9"}, Pattern: 1},
        Border: []excelize.Border{
            {Type: "left", Color: "A5D6A7", Style: 1},
            {Type: "right", Color: "A5D6A7", Style: 1},
            {Type: "top", Color: "A5D6A7", Style: 1},
            {Type: "bottom", Color: "A5D6A7", Style: 1},
        },
    })
    expenseRowEven, _ := f.NewStyle(&excelize.Style{
        Fill: excelize.Fill{Type: "pattern", Color: []string{"FFC7CE"}, Pattern: 1},
        Border: []excelize.Border{
            {Type: "left", Color: "EF9A9A", Style: 1},
            {Type: "right", Color: "EF9A9A", Style: 1},
            {Type: "top", Color: "EF9A9A", Style: 1},
            {Type: "bottom", Color: "EF9A9A", Style: 1},
        },
    })
    expenseRowOdd, _ := f.NewStyle(&excelize.Style{
        Fill: excelize.Fill{Type: "pattern", Color: []string{"FFCDD2"}, Pattern: 1},
        Border: []excelize.Border{
            {Type: "left", Color: "EF9A9A", Style: 1},
            {Type: "right", Color: "EF9A9A", Style: 1},
            {Type: "top", Color: "EF9A9A", Style: 1},
            {Type: "bottom", Color: "EF9A9A", Style: 1},
        },
    })

    // ===== CÁLCULOS =====
    var totalIncome, totalExpense float64
    incomeByCat := make(map[string]float64)
    expenseByCat := make(map[string]float64)
    for _, e := range filtered {
        if e.Type == "Entrada" {
            totalIncome += e.Amount
            incomeByCat[e.Category] += e.Amount
        } else {
            totalExpense += e.Amount
            expenseByCat[e.Category] += e.Amount
        }
    }
    balance := totalIncome - totalExpense

    // ===== HOJA 1: RESUMEN GENERAL =====
    f.SetSheetName("Sheet1", "Resumen")
    f.SetCellValue("Resumen", "A1", fmt.Sprintf("REPORTE MENSUAL - %s %s", month, year))
    f.MergeCell("Resumen", "A1", "C1")
    f.SetCellStyle("Resumen", "A1", "C1", titleStyle)
    f.SetCellValue("Resumen", "A2", fmt.Sprintf("Moneda: %s", currency))
    f.MergeCell("Resumen", "A2", "C2")
    f.SetCellStyle("Resumen", "A2", "C2", subtitleStyle)
    f.SetCellValue("Resumen", "A4", "Concepto")
    f.SetCellValue("Resumen", "B4", "Tipo")
    f.SetCellValue("Resumen", "C4", "Monto")
    f.SetCellStyle("Resumen", "A4", "C4", headerStyle)
    f.SetCellValue("Resumen", "A5", "Total Entradas")
    f.SetCellValue("Resumen", "B5", "Ingreso")
    f.SetCellValue("Resumen", "C5", formatCurrency(totalIncome, currency))
    f.SetCellStyle("Resumen", "A5", "C5", incomeStyle)
    f.SetCellValue("Resumen", "A6", "Total Gastos")
    f.SetCellValue("Resumen", "B6", "Egreso")
    f.SetCellValue("Resumen", "C6", formatCurrency(totalExpense, currency))
    f.SetCellStyle("Resumen", "A6", "C6", expenseStyle)
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
    f.SetColWidth("Resumen", "A", "A", 25)
    f.SetColWidth("Resumen", "B", "B", 15)
    f.SetColWidth("Resumen", "C", "C", 20)

    // ===== HOJA 2: POR CATEGORÍAS =====
    f.NewSheet("Por Categorías")
    f.SetCellValue("Por Categorías", "A1", "RESUMEN POR CATEGORÍAS")
    f.MergeCell("Por Categorías", "A1", "C1")
    f.SetCellStyle("Por Categorías", "A1", "C1", titleStyle)
    f.SetCellValue("Por Categorías", "A3", "Categoría")
    f.SetCellValue("Por Categorías", "B3", "Tipo")
    f.SetCellValue("Por Categorías", "C3", "Monto")
    f.SetCellStyle("Por Categorías", "A3", "C3", headerStyle)

    row := 4
    if len(incomeOrder) > 0 {
        f.SetCellValue("Por Categorías", fmt.Sprintf("A%d", row), "ENTRADAS")
        f.MergeCell("Por Categorías", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row))
        f.SetCellStyle("Por Categorías", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), subtitleStyle)
        row++
        for _, cat := range incomeOrder {
            val, exists := incomeByCat[cat]
            if !exists { continue }
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
        f.SetCellValue("Por Categorías", fmt.Sprintf("A%d", row), "Total Entradas")
        f.SetCellValue("Por Categorías", fmt.Sprintf("C%d", row), formatCurrency(totalIncome, currency))
        f.SetCellStyle("Por Categorías", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), incomeStyle)
        row += 2
    }

    if len(expenseOrder) > 0 {
        f.SetCellValue("Por Categorías", fmt.Sprintf("A%d", row), "GASTOS")
        f.MergeCell("Por Categorías", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row))
        f.SetCellStyle("Por Categorías", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), subtitleStyle)
        row++
        for _, cat := range expenseOrder {
            val, exists := expenseByCat[cat]
            if !exists { continue }
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
        f.SetCellValue("Por Categorías", fmt.Sprintf("A%d", row), "Total Gastos")
        f.SetCellValue("Por Categorías", fmt.Sprintf("C%d", row), formatCurrency(totalExpense, currency))
        f.SetCellStyle("Por Categorías", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), expenseStyle)
    }
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

    // ===== HOJAS 3 Y 4: ENTRADAS / GASTOS =====
    incomeEntries := filterAndGroupByCategory(filtered, "Entrada", incomeOrder)
    expenseEntries := filterAndGroupByCategory(filtered, "Gasto", expenseOrder)

	createDetailedSheet(f, "Entradas", incomeEntries, incomeRowEven, incomeRowOdd, headerStyle, titleStyle, incomeStyle, expenseStyle, totalIncome, currency)
	createDetailedSheet(f, "Gastos", expenseEntries, expenseRowEven, expenseRowOdd, headerStyle, titleStyle, incomeStyle, expenseStyle, totalExpense, currency)
    f.SetActiveSheet(0)
    return f.Write(w)
}

// parseAmountAndCurrency procesa una cadena como "200,000 CUP" o "1500.50 EUR"
// y devuelve el valor numérico y el código de moneda (si lo encuentra).
func parseAmountAndCurrency(s string) (float64, string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, "", fmt.Errorf("vacío")
	}
	// Separar por espacio: la última parte que coincida con un código de moneda conocido
	parts := strings.Split(s, " ")
	currency := ""
	numPart := s
	// Buscar moneda conocida en cualquier parte final
	for i := len(parts) - 1; i >= 0; i-- {
		upper := strings.ToUpper(strings.TrimSpace(parts[i]))
		if upper == "USD" || upper == "EUR" || upper == "CUP" {
			currency = upper
			numPart = strings.Join(parts[:i], " ")
			break
		}
	}
	// Si no encontramos moneda, todo es número
	if currency == "" {
		numPart = s
	}

	amount, err := parseAmount(numPart) // reutiliza la función parseAmount existente
	return amount, currency, err
}

// ImportEntries lee un archivo Excel con el formato de guardado
func ImportEntries(path string) ([]model.Entry, error) {
    f, err := excelize.OpenFile(path)
    if err != nil {
        return nil, fmt.Errorf("no se pudo abrir el archivo: %v", err)
    }
    defer f.Close()

    var rows [][]string
    // Intentar leer hoja "Listado" (formato antiguo)
    rows, err = f.GetRows("Listado", excelize.Options{RawCellValue: true})
    if err != nil {
        // Si no existe, leer "Entradas" y "Gastos"
        rowsEnt, err1 := f.GetRows("Entradas", excelize.Options{RawCellValue: true})
        rowsGas, err2 := f.GetRows("Gastos", excelize.Options{RawCellValue: true})
        if err1 != nil || err2 != nil {
            // Si falla, intentar con la primera hoja
            sheets := f.GetSheetList()
            if len(sheets) > 0 {
                rows, err = f.GetRows(sheets[0], excelize.Options{RawCellValue: true})
                if err != nil {
                    return nil, fmt.Errorf("no se pudieron leer los datos: %v", err)
                }
            } else {
                return nil, fmt.Errorf("el archivo no contiene hojas válidas")
            }
        } else {
            // Combinar filas de ambas hojas (omitiendo cabeceras y filas de subcategorías)
            rows = append(rows, rowsEnt[0]) // cabecera (no se usa pero para mantener índices)
            rows = append(rows, extractDataRows(rowsEnt)...)
            rows = append(rows, extractDataRows(rowsGas)...)
        }
    } else {
        // Ya tenemos rows de "Listado"
    }

    var entries []model.Entry
    for i, row := range rows {
        if i == 0 { continue }
        if len(row) < 6 { continue }
        // Si es una fila de subcategoría (celda A contiene texto sin fecha), la ignoramos
        if _, err := time.Parse("02/01/2006", row[0]); err != nil {
            continue
        }
        date, _ := time.Parse("02/01/2006", row[0])
        amount, extractedCurrency, err := parseAmountAndCurrency(row[5])
        if err != nil { amount = 0 }
        currency := "USD"
        if len(row) > 6 && strings.TrimSpace(row[6]) != "" {
            currency = strings.TrimSpace(row[6])
        } else if extractedCurrency != "" {
            currency = extractedCurrency
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

// extractDataRows extrae las filas de datos de una hoja, saltando cabeceras y filas de título.
func extractDataRows(allRows [][]string) [][]string {
    var data [][]string
    for _, row := range allRows {
        if len(row) >= 6 && row[0] != "" && row[0] != "Fecha" {
            // comprobar si es una fila de subcategoría (una sola celda larga)
            if len(row) == 1 || (len(row) > 1 && row[1] == "" && row[2] == "") {
                continue
            }
            data = append(data, row)
        }
    }
    return data
}

func parseAmount(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("vacío")
	}

	// Quitar códigos de moneda y espacios internos
	for _, code := range []string{"USD", "EUR", "CUP"} {
		s = strings.ReplaceAll(s, code, "")
	}
	s = strings.ReplaceAll(s, " ", "")
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("vacío")
	}

	lastComma := strings.LastIndex(s, ",")
	lastDot := strings.LastIndex(s, ".")

	switch {
	case lastComma >= 0 && lastDot >= 0:
		if lastComma > lastDot {
			// Formato español: 1.234,56
			s = strings.ReplaceAll(s, ".", "")
			s = strings.ReplaceAll(s, ",", ".")
		} else {
			// Formato inglés / nuevo: 1,234.56
			s = strings.ReplaceAll(s, ",", "")
		}
	case lastComma >= 0:
		// Solo coma: si hay 3 dígitos detrás → separador de miles
		if len(s)-lastComma-1 == 3 {
			s = strings.ReplaceAll(s, ",", "")
		} else {
			s = strings.Replace(s, ",", ".", 1)
		}
	case lastDot >= 0:
		// Solo punto: si hay 3 dígitos detrás → separador de miles (legado)
		if len(s)-lastDot-1 == 3 && strings.Count(s, ".") > 0 {
			s = strings.ReplaceAll(s, ".", "")
		}
	}

	return strconv.ParseFloat(s, 64)
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

func formatCurrency(amount float64, currency string) string {
	str := strconv.FormatFloat(amount, 'f', 2, 64)
	neg := false
	if strings.HasPrefix(str, "-") {
		neg = true
		str = str[1:]
	}
	parts := strings.Split(str, ".")
	intPart, decPart := parts[0], parts[1]

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

// filterAndGroupByCategory filtra las entradas por tipo y las agrupa en el orden dado.
// Devuelve una lista de pares (categoría, lista de entradas).
func filterAndGroupByCategory(entries []model.Entry, tipo string, order []string) [][]model.Entry {
    // Agrupar
    groups := make(map[string][]model.Entry)
    for _, e := range entries {
        if e.Type == tipo {
            groups[e.Category] = append(groups[e.Category], e)
        }
    }
    // Ordenar según order, y luego por fecha dentro de cada grupo
    result := make([][]model.Entry, 0, len(order))
    for _, cat := range order {
        if list, ok := groups[cat]; ok {
            // Ordenar por fecha ascendente
            sort.Slice(list, func(i, j int) bool {
                return list[i].Date.Before(list[j].Date)
            })
            result = append(result, list)
        }
    }
    // Añadir categorías nuevas que no estén en order (al final)
    for cat, list := range groups {
        found := false
        for _, oc := range order {
            if oc == cat {
                found = true
                break
            }
        }
        if !found {
            sort.Slice(list, func(i, j int) bool {
                return list[i].Date.Before(list[j].Date)
            })
            result = append(result, list)
        }
    }
    return result
}

// createDetailedSheet crea una hoja con todas las transacciones de un tipo,
// agrupadas por categoría.
func createDetailedSheet(f *excelize.File, sheetName string,
    grouped [][]model.Entry, evenStyle, oddStyle, headerStyle, titleStyle,
    incomeStyle, expenseStyle int, total float64, currency string) {

    f.NewSheet(sheetName)
    f.SetCellValue(sheetName, "A1", fmt.Sprintf("RESUMEN DE %s", strings.ToUpper(sheetName)))
    f.MergeCell(sheetName, "A1", "G1")
    f.SetCellStyle(sheetName, "A1", "G1", titleStyle)

    headers := []string{"Fecha", "Tipo", "Categoría", "# Factura", "Descripción", "Monto", "Moneda"}
    f.SetCellStyle(sheetName, "A3", "G3", headerStyle)
    for i, h := range headers {
        cell := fmt.Sprintf("%c3", 'A'+i)
        f.SetCellValue(sheetName, cell, h)
    }

    row := 4
    for _, group := range grouped {
        if len(group) == 0 { continue }
        category := group[0].Category
        f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), category)
        f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row))
        subStyle, _ := f.NewStyle(&excelize.Style{
            Font:      &excelize.Font{Bold: true, Color: "2E75B6"},
            Fill:      excelize.Fill{Type: "pattern", Color: []string{"E8F0FE"}, Pattern: 1},
            Alignment: &excelize.Alignment{Horizontal: "left"},
        })
        f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), subStyle)
        row++

        for i, e := range group {
            f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), e.Date.Format("02/01/2006"))
            f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), e.Type)
            f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), e.Category)
            f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), e.InvoiceNum)
            f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), e.Description)
            f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), e.Amount)
            f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), e.Currency)

            var styleID int
            if i%2 == 0 {
                styleID = evenStyle
            } else {
                styleID = oddStyle
            }
            f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), styleID)
            row++
        }
        row++ // espacio entre categorías
    }

    // Fila de total
    f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "TOTAL")
    f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("E%d", row))
    f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), formatCurrency(total, currency))
    if sheetName == "Entradas" {
        f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), incomeStyle)
    } else {
        f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), expenseStyle)
    }

    f.SetColWidth(sheetName, "A", "A", 12)
    f.SetColWidth(sheetName, "B", "B", 10)
    f.SetColWidth(sheetName, "C", "C", 20)
    f.SetColWidth(sheetName, "D", "D", 12)
    f.SetColWidth(sheetName, "E", "E", 40)
    f.SetColWidth(sheetName, "F", "F", 15)
    f.SetColWidth(sheetName, "G", "G", 10)
}