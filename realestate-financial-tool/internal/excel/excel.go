package excel

import (
	"database/sql"
	"fmt"

	// ...existing imports...
	"github.com/xuri/excelize/v2"
)

// ExportToExcel exports SQL query results to an Excel file
func ExportToExcel(db *sql.DB, query string, filename string) error {
	// Execute the query
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// Create a new Excel file
	f := excelize.NewFile()
	defer f.Close()

	// Set column headers
	for i, col := range columns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Sheet1", cell, col)
	}

	// Write data rows
	rowIndex := 2
	for rows.Next() {
		// Create a slice of interface{} to store the row
		values := make([]interface{}, len(columns))
		valuePointers := make([]interface{}, len(columns))
		for i := range values {
			valuePointers[i] = &values[i]
		}

		// Scan the row into the slice
		if err := rows.Scan(valuePointers...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Write each cell in the row
		for i := range values {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowIndex)
			f.SetCellValue("Sheet1", cell, values[i])
		}
		rowIndex++
	}

	// Save the file
	if err := f.SaveAs(filename); err != nil {
		return fmt.Errorf("failed to save excel file: %w", err)
	}

	return nil
}

// // Example usage in your remote() function:
// func remote() {
// 	// ...existing code...

// 	db := sql.OpenDB(connector)
// 	defer db.Close()

// 	// Example: Export query results to Excel
// 	err = ExportToExcel(db, "SELECT * FROM your_table", "output.xlsx")
// 	if err != nil {
// 		fmt.Printf("Error exporting to Excel: %v\n", err)
// 		os.Exit(1)
// 	}
// }
