package utils

import (
	"fmt"
	"time"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/xuri/excelize/v2"
)

func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// 辅助函数：判断DataFrame是否有某列
func HasColumn(df dataframe.DataFrame, name string) bool {
	for _, n := range df.Names() {
		if n == name {
			return true
		}
	}
	return false
}

func ParseTime(s series.Element) (time.Time, error) {
	if s.String() == "" || s.IsNA() {
		return time.Time{}, nil
	}
	t, err := time.Parse("2006-01-02 15:04:05", s.String())
	if err != nil {
		return time.Time{}, err // 返回零值时间表示解析失败
	}
	return t, err
}

func SubSeriesTime(df dataframe.DataFrame, colName1, colName2, colName3 string) (dataframe.DataFrame, error) {

	// 获取两列的所有元素
	col1 := df.Col(colName1)
	col2 := df.Col(colName2)

	// 预分配切片容量
	durations := make([]float64, 0, df.Nrow())

	// 遍历每一行计算时间差
	for i := 0; i < df.Nrow(); i++ {
		startTime, err := ParseTime(col1.Elem(i))
		if err != nil {
			return df, fmt.Errorf("failed to parse start time at row %d: %v", i, err)
		}

		endTime, err := ParseTime(col2.Elem(i))
		if err != nil {
			return df, fmt.Errorf("failed to parse end time at row %d: %v", i, err)
		}

		duration := startTime.Sub(endTime).Seconds()
		durations = append(durations, duration)
	}

	// 创建时间差列并添加到DataFrame
	durationCol := series.New(durations, series.Float, colName3)

	// 使用指针更新DataFrame
	return df.CBind(dataframe.New(durationCol)), nil
}

func SaveToExcel(df dataframe.DataFrame, filePath string) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"

	// 写入列名
	colNames := df.Names()
	for i, name := range colNames {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, name)
	}

	// 写入数据
	for rowIdx := 0; rowIdx < df.Nrow(); rowIdx++ {
		for colIdx, colName := range colNames {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			val := df.Col(colName).Val(rowIdx)
			f.SetCellValue(sheetName, cell, val)
		}
	}

	// 保存文件
	if err := f.SaveAs(filePath); err != nil {
		return fmt.Errorf("保存Excel文件失败: %w", err)
	}

	fmt.Printf("\n处理后的数据已保存到: %s\n", filePath)
	return nil
}
