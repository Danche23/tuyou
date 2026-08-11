package excel

import (
	"strconv"

	"github.com/xuri/excelize/v2"
)

// Build 生成单 Sheet 的 xlsx 二进制流。
// headers 为表头行，rows 每项对应一行数据，顺序须与 headers 一致。
func Build(sheetName string, headers []string, rows [][]interface{}) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	if sheetName == "" {
		sheetName = "Sheet1"
	}
	f.SetSheetName("Sheet1", sheetName)

	for i, h := range headers {
		cell := columnName(i+1) + "1"
		if err := f.SetCellValue(sheetName, cell, h); err != nil {
			return nil, err
		}
	}

	for idx, row := range rows {
		rowNum := idx + 2
		for i, v := range row {
			cell := columnName(i+1) + strconv.Itoa(rowNum)
			if err := f.SetCellValue(sheetName, cell, v); err != nil {
				return nil, err
			}
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// columnName 生成 Excel 列名（A、B、...、Z、AA...）
func columnName(col int) string {
	name := ""
	for col > 0 {
		col--
		name = string(rune('A'+col%26)) + name
		col /= 26
	}
	return name
}
