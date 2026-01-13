package services

import (
	"scandata/models"

	"github.com/xuri/excelize/v2"
)

func GenerateExcel(scans []models.ScanLog) (*excelize.File, error) {
	f := excelize.NewFile()
	sheetName := "Scan Report"
	f.SetSheetName("Sheet1", sheetName)

	// Set headers
	headers := []string{"No", "Tanggal", "Waktu", "Unit", "QR Code", "Lokasi", "Scanner", "Sesuai", "Catatan"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Style for header
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	f.SetRowStyle(sheetName, 1, 1, headerStyle)

	// Add data
	for i, scan := range scans {
		row := i + 2

		isMatchText := "Tidak Sesuai"
		if scan.IsMatch {
			isMatchText = "Sesuai"
		}

		f.SetCellValue(sheetName, cellName(1, row), i+1)
		f.SetCellValue(sheetName, cellName(2, row), scan.ScannedAt.Format("2006-01-02"))
		f.SetCellValue(sheetName, cellName(3, row), scan.ScannedAt.Format("15:04:05"))
		f.SetCellValue(sheetName, cellName(4, row), scan.Unit.Name)
		f.SetCellValue(sheetName, cellName(5, row), scan.Unit.QRCode)
		f.SetCellValue(sheetName, cellName(6, row), scan.Unit.Location)
		f.SetCellValue(sheetName, cellName(7, row), scan.User.Name)
		f.SetCellValue(sheetName, cellName(8, row), isMatchText)
		f.SetCellValue(sheetName, cellName(9, row), scan.Notes)
	}

	// Auto adjust column width
	for i := 1; i <= 9; i++ {
		colName, _ := excelize.ColumnNumberToName(i)
		f.SetColWidth(sheetName, colName, colName, 15)
	}

	return f, nil
}

func cellName(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}
