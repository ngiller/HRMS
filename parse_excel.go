package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"strings"
)

func main() {
	f, err := excelize.OpenFile("../employee_template.xlsx")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		fmt.Println("Error reading rows:", err)
		return
	}

	if len(rows) < 2 {
		fmt.Println("Not enough rows")
		return
	}

	headers := rows[0]
	approvalColIndex := -1
	for i, h := range headers {
		if strings.Contains(strings.ToLower(h), "approval") {
			approvalColIndex = i
			fmt.Printf("Found approval column: %s at index %d\n", h, i)
		}
	}

	if approvalColIndex == -1 {
		approvalColIndex = 39 // 0-indexed column 39 (row 40) based on my previous mapping
		fmt.Printf("Fallback to column index %d: %s\n", approvalColIndex, headers[approvalColIndex])
	}

	uniqueApprovals := make(map[string]bool)
	for _, row := range rows[1:] {
		if approvalColIndex < len(row) {
			val := strings.TrimSpace(row[approvalColIndex])
			if val != "" {
				uniqueApprovals[val] = true
			}
		}
	}

	fmt.Println("\nUnique Approval Lines:")
	for k := range uniqueApprovals {
		fmt.Printf("- %s\n", k)
	}
}
