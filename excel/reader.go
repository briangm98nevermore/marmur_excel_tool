package excel

import (
	"github.com/xuri/excelize/v2"
)

// ReadAllRows lee todas las filas de una hoja específica.
func ReadAllRows(filePath, sheetName string) ([][]string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	return rows, nil
}