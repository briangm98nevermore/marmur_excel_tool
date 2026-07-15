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
	// Filtrar por moneda (por si hay varias, aunque la UI solo deja una)
	var filtered []model.Entry
	for _, e := range entries {
		if e.Currency == currency {
			filtered = append(filtered, e)
		}
	}
	if len(filtered) == 0 {
		return fmt.Errorf("no hay datos para la moneda %s", currency)
	}

	// Hoja de resumen general
	f.SetSheetName("Sheet1", "Resumen")
	f.SetCellValue("Resumen", "A1", fmt.Sprintf("Reporte %s %s - %s", month, year, currency))

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

	f.SetCellValue("Resumen", "A3", "Total Entradas")
	f.SetCellValue("Resumen", "B3", totalIncome)
	f.SetCellValue("Resumen", "A4", "Total Gastos")
	f.SetCellValue("Resumen", "B4", totalExpense)
	f.SetCellValue("Resumen", "A5", "Balance")
	f.SetCellValue("Resumen", "B5", balance)

	// Hoja de detalle por nomenclador
	incomeByCat := make(map[string]float64)
	expenseByCat := make(map[string]float64)
	for _, e := range filtered {
		if e.Type == "Entrada" {
			incomeByCat[e.Category] += e.Amount
		} else {
			expenseByCat[e.Category] += e.Amount
		}
	}
	// Crear hoja "Por Categorías"
	f.NewSheet("Por Categorías")
	row := 1
	f.SetCellValue("Por Categorías", "A1", "Categoría")
	f.SetCellValue("Por Categorías", "B1", "Tipo")
	f.SetCellValue("Por Categorías", "C1", "Monto")
	row++
	for cat, val := range incomeByCat {
		f.SetCellValue("Por Categorías", fmt.Sprintf("A%d", row), cat)
		f.SetCellValue("Por Categorías", fmt.Sprintf("B%d", row), "Entrada")
		f.SetCellValue("Por Categorías", fmt.Sprintf("C%d", row), val)
		row++
	}
	for cat, val := range expenseByCat {
		f.SetCellValue("Por Categorías", fmt.Sprintf("A%d", row), cat)
		f.SetCellValue("Por Categorías", fmt.Sprintf("B%d", row), "Gasto")
		f.SetCellValue("Por Categorías", fmt.Sprintf("C%d", row), val)
		row++
	}

	// Hoja de listado completo
	f.NewSheet("Listado")
	f.SetCellValue("Listado", "A1", "Fecha")
	f.SetCellValue("Listado", "B1", "Tipo")
	f.SetCellValue("Listado", "C1", "Categoría")
	f.SetCellValue("Listado", "D1", "# Factura")
	f.SetCellValue("Listado", "E1", "Descripción")
	f.SetCellValue("Listado", "F1", "Monto")
	for i, e := range filtered {
		row := i + 2
		f.SetCellValue("Listado", fmt.Sprintf("A%d", row), e.Date.Format("02/01/2006"))
		f.SetCellValue("Listado", fmt.Sprintf("B%d", row), e.Type)
		f.SetCellValue("Listado", fmt.Sprintf("C%d", row), e.Category)
		f.SetCellValue("Listado", fmt.Sprintf("D%d", row), e.InvoiceNum)
		f.SetCellValue("Listado", fmt.Sprintf("E%d", row), e.Description)
		f.SetCellValue("Listado", fmt.Sprintf("F%d", row), e.Amount)
	}

	// Eliminar hoja por defecto si es necesario
	f.DeleteSheet("Sheet1")
	f.SetActiveSheet(0) // Resumen

	return f.Write(w)
}

// ImportEntries lee un archivo Excel con el formato de guardado
func ImportEntries(path string) ([]model.Entry, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := f.GetRows("Listado") // Esperamos la hoja "Listado"
	if err != nil {
		return nil, err
	}
	var entries []model.Entry
	for i, row := range rows {
		if i == 0 {
			continue // cabecera
		}
		if len(row) < 6 {
			continue
		}
		date, _ := time.Parse("02/01/2006", row[0])
		amount, _ := strconv.ParseFloat(row[5], 64)
		entries = append(entries, model.Entry{
			Date:        date,
			Type:        row[1],
			Category:    row[2],
			InvoiceNum:  row[3],
			Description: row[4],
			Amount:      amount,
			Currency:    "USD", // Asume USD, podrías guardar la moneda en otra columna
		})
	}
	return entries, nil
}

// SaveEntries guarda las entradas en un Excel intermedio
func SaveEntries(entries []model.Entry, w io.Writer) error {
	f := excelize.NewFile()
	f.SetSheetName("Sheet1", "Listado")
	f.SetCellValue("Listado", "A1", "Fecha")
	f.SetCellValue("Listado", "B1", "Tipo")
	f.SetCellValue("Listado", "C1", "Categoría")
	f.SetCellValue("Listado", "D1", "# Factura")
	f.SetCellValue("Listado", "E1", "Descripción")
	f.SetCellValue("Listado", "F1", "Monto")
	f.SetCellValue("Listado", "G1", "Moneda")
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
	return f.Write(w)
}