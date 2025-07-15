// handler.go
package email

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/tealeg/xlsx"
	"github.com/xuri/excelize/v2"
)

const (
	Number   string = "[0-9.]+"
	InString string = "[\\D]+"
)

// DataFrameWrapper 封装DataFrame并提供线程安全访问
type DataFrameWrapper struct {
	df dataframe.DataFrame // 存储DataFrame数据
	mu sync.RWMutex        // 读写锁保证线程安全
}

// GetDF 获取当前DataFrame(线程安全)
func (d *DataFrameWrapper) GetDF() dataframe.DataFrame {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.df
}

// GetDF 获取当前DataFrame(线程安全)
func (d *DataFrameWrapper) SetDF(df dataframe.DataFrame) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.df = df
}

// LoadXLSX 从XLSX数据加载到DataFrameZ
func (d *DataFrameWrapper) ReadXLSX(data []byte, sheetName string) error {

	// 1. 使用tealeg/xlsx打开Excel文件
	xlFile, err := xlsx.OpenBinary(data)
	if err != nil {
		return err
	}

	// 2. 获取第一个工作表
	if len(xlFile.Sheets) == 0 {
		return fmt.Errorf("excel文件中没有工作表: %w", err)
	}
	sheet := xlFile.Sheet[sheetName]

	// 3. 转换为Gota DataFrame
	if err := d.convertSheetToDataFrame(sheet); err != nil {
		return fmt.Errorf("转换为dataframe失败")
	}

	// 4. 初始化id
	d.addFlightIndex()

	// 5. 格式化数据
	d.ProcessData()

	return err
}

// convertSheetToDataFrame 将xlsx.Sheet转换为dataframe.DataFrame
func (d *DataFrameWrapper) convertSheetToDataFrame(sheet *xlsx.Sheet) error {
	if len(sheet.Rows) == 0 {
		return fmt.Errorf("sheet rows wei 0")
	}

	// 获取列名(假设第二行是标题行)
	var headers []string
	for _, cell := range sheet.Rows[1].Cells {
		headers = append(headers, cell.String())
	}

	// 准备数据列
	columns := make([][]string, len(headers))
	for i := range columns {
		columns[i] = make([]string, 0, len(sheet.Rows)-1)
	}

	// 填充数据(从第三行开始)
	for _, row := range sheet.Rows[2:] {
		for i, cell := range row.Cells {
			if i < len(headers) { // 确保不超出列数范围
				columns[i] = append(columns[i], cell.String())
			}
		}
	}

	// 创建Series切片
	seriesList := make([]series.Series, len(headers))
	for i, colName := range headers {
		// 自动推断类型创建Series

		seriesList[i] = series.New(columns[i], series.String, colName)
	}

	d.SetDF(dataframe.New(seriesList...))
	return nil
}

// saveToExcel 将DataFrame保存回Excel文件
func (d *DataFrameWrapper) SaveToExcel(filePath string) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"

	df := d.df
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

func (d *DataFrameWrapper) addFlightIndex() {
	df := d.df
	// 1. 拼接航班号 + 计划时间
	combined := df.Col("离港航班号").Records()[:] // 跳过列名
	scheduledTimes := df.Col("计划撤轮挡时间").Records()[:]
	for i := 0; i < len(combined); i++ {
		combined[i] += scheduledTimes[i] // 拼接字符串
	}

	// 2. 计算 MD5 哈希
	md5Hashes := make([]string, len(combined))
	for i, s := range combined {
		hash := md5.Sum([]byte(s))
		md5Hashes[i] = hex.EncodeToString(hash[:])
	}

	// 3. 将 MD5 结果添加到 DataFrame
	df = df.Mutate(
		series.New(md5Hashes, series.String, "flightId"),
	)
	d.SetDF(df)
}

// ProcessData 处理DataFrame数据
func (d *DataFrameWrapper) ProcessData() error {
	df := d.df

	// 1. 去除完全空的行
	// df = dfPtr.Filter(
	// 	dataframe.F{Colname: dfPtr.Names()[0], Comparator: series.Neq, Comparando: ""},
	// )

	// 2. 标准化时间字段
	timeCols := d.findTimeColumns()

	for _, col := range timeCols {
		df = df.Mutate(
			series.New(df.Col(col).Map(excelToTime), series.String, col),
		)
	}

	// 更新处理后的数据
	d.SetDF(df)

	return nil
}

// 辅助函数：查找可能是时间类型的列
func (d *DataFrameWrapper) findTimeColumns() []string {
	var timeCols []string
	timeKeywords := []string{"时间", "日期", "date", "time", "COBT"}

	for _, col := range d.df.Names() {
		for _, kw := range timeKeywords {
			if strings.Contains(col, kw) {
				timeCols = append(timeCols, col)
				break
			}
		}
	}
	return timeCols
}

// excel时间类型转time.Time类型
func excelToTime(v series.Element) series.Element {
	re := regexp.MustCompile(Number)

	// 1. 错误处理：检查元素是否为数值类型
	if !isMatchGood(re, v.String()) {
		return v // 返回原值或可设置为错误标记
	}

	// 2. 处理Excel的1900年闰年错误（2月29日不存在）
	excelDays := v.Float()
	if excelDays >= 60 {
		excelDays -= 1 // 调整60天后的日期
	}

	// 3. 优化时间计算（减少临时变量）
	base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	days := int(excelDays)
	fraction := excelDays - float64(days)

	// 4. 更精确的时间计算（包含纳秒）
	result := base.AddDate(0, 0, days).
		Add(time.Duration(86400*fraction*1e9) * time.Nanosecond)

	// 5. 保留原始时间值并设置格式化字符串
	res := result.Format("2006-01-02 15:04:05")

	resVO := reflect.ValueOf(res)
	v.Set(resVO.Interface())

	return v
}

// 正确做法 - 预编译
func isMatchGood(re *regexp.Regexp, s string) bool {
	return re.MatchString(s)
}
