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

func SubSeriesTime(df *dataframe.DataFrame, colName1, colName2, colName3 string) error {
	// 1. 校验输入列是否存在
	if !HasColumn(*df, colName1) || !HasColumn(*df, colName2) {
		return fmt.Errorf("列 %s 或 %s 不存在", colName1, colName2)
	}

	// 2. 提前获取列和行数，避免重复计算
	col1, col2 := df.Col(colName1), df.Col(colName2)
	rowCount := df.Nrow()
	durations := make([]float64, rowCount) // 预分配精确大小

	// 3. 遍历计算时间差（秒）
	for i := 0; i < rowCount; i++ {
		startTime, err := ParseTime(col1.Elem(i))
		if err != nil {
			return fmt.Errorf("第 %d 行解析开始时间失败: %v", i, err)
		}
		endTime, err := ParseTime(col2.Elem(i))
		if err != nil {
			return fmt.Errorf("第 %d 行解析结束时间失败: %v", i, err)
		}
		durations[i] = startTime.Sub(endTime).Seconds()
	}

	// 4. 直接使用 series.New 创建新列并添加
	durationSeries := series.New(durations, series.Float, colName3)
	df.CBind(dataframe.New(durationSeries))

	return nil
}

func NewSeries[T1 comparable](colName string, value T1, seriesType series.Type, nrow int) (series.Series, error) {

	durations := make([]T1, nrow)
	for i := 0; i < nrow; i++ {
		durations[i] = value
	}
	durationSeries := series.New(durations, seriesType, colName)
	return durationSeries, nil
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
