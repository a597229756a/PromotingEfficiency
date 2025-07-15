// reader.go
package file

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/tealeg/xlsx"
	"github.com/xuri/excelize/v2"
)

func ReadXLSXToDataFrame(filePath, sheetName string) (dataframe.DataFrame, error) {
	df, err := ReadXLSX(filePath, sheetName)
	if err != nil {
		return dataframe.DataFrame{}, fmt.Errorf("failed to open xlsx file: %w", err)
	}

	return df, nil
}

const (
	Number   string = "[0-9.]+"
	InString string = "[\\D]+"
)

// Config 配置结构体
type Config struct {
	FolderName    string
	Level         int
	Keyword       string
	CheckInterval time.Duration
}

// FileInfo 文件信息结构体
type FileInfo struct {
	Name     string
	FullPath string
	ModTime  time.Time
	df       *dataframe.DataFrame
	mu       sync.RWMutex
}

// ensureDir 确保目录存在
func ensureDir(dirPath string) error {
	if info, err := os.Stat(dirPath); err == nil {
		if info.IsDir() {
			return nil
		}
		return fmt.Errorf("%s exists but is not a directory", dirPath)
	}
	return os.MkdirAll(dirPath, 0755)
}

// getTargetFolder 获取目标文件夹路径
func GetTargetFolder(folderName string, level int) (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	path := exePath
	for i := 0; i < level; i++ {
		path = filepath.Dir(path)
	}

	return filepath.Join(path, folderName), nil
}

// // findLatestExcel 查找最新的符合条件的Excel文件
// func (fm *FileMonitor) findLatestExcel() (*FileInfo, error) {
// 	entries, err := os.ReadDir(fm.targetDir)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read directory: %w", err)
// 	}

// 	var latest *FileInfo

// 	for _, entry := range entries {
// 		if entry.IsDir() {
// 			continue
// 		}

// 		info, err := entry.Info()
// 		if err != nil {
// 			continue
// 		}

// 		if !strings.EqualFold(filepath.Ext(info.Name()), ".xlsx") ||
// 			!strings.Contains(info.Name(), fm.config.Keyword) {
// 			continue
// 		}

// 		fullPath := filepath.Join(fm.targetDir, info.Name())

// 		if latest == nil || info.ModTime().After(latest.ModTime) {

// 			t1 := time.Now()
// 			df, err := ReadXLSX(fullPath, "进离港航班")
// 			fmt.Println(time.Since(t1))

// 			if err != nil {
// 				return nil, fmt.Errorf("sheet name 进离港航班 获取失败 %w", err)
// 			}
// 			latest = &FileInfo{
// 				Name:     info.Name(),
// 				FullPath: fullPath,
// 				ModTime:  info.ModTime(),
// 				df:       &df,
// 			}
// 		}
// 	}

// 	if latest == nil {
// 		return nil, fmt.Errorf("no matching .xlsx files found")
// 	}

// 	return latest, nil
// }

// // Run 启动文件监控
// func (fm *FileMonitor) Run(ctx context.Context) error {
// 	// 初始检查
// 	latest, err := fm.findLatestExcel()
// 	if err != nil {
// 		return fmt.Errorf("initial file check failed: %w", err)
// 	}
// 	fm.updateLastFile(latest)
// 	fm.printFileInfo(latest)

// 	// 设置定时器
// 	ticker := time.NewTicker(fm.config.CheckInterval)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return nil
// 		case <-ticker.C:
// 			latest, err := fm.findLatestExcel()
// 			if err != nil {
// 				fmt.Printf("Monitoring error: %v\n", err)
// 				continue
// 			}

// 			if fm.isNewFile(latest) {
// 				fm.updateLastFile(latest)
// 				fm.printFileInfo(latest)

// 				// 对数据进行数据清理
// 				if err := latest.ProcessData(); err != nil {
// 					fmt.Printf("Data processing error: %v\n", err)
// 					continue
// 				}

// 				// 对数据列进行计算需要修改file
// 				latest.addFlightIndex()

// 				// latest.SaveToExcel("DATA.xlsx")
// 				hd := handle.DelayProcessor{}
// 				hd.UpdateDelay(latest.df)
// 				// // 可以在这里添加数据分析或存储逻辑
// 				// fm.analyzeData(latest.df)
// 			}
// 		}
// 	}
// }

// func (fm *FileMonitor) isNewFile(file *FileInfo) bool {
// 	fm.mu.Lock()
// 	defer fm.mu.Unlock()

// 	return fm.lastFile == nil ||
// 		file.ModTime.After(fm.lastFile.ModTime) ||
// 		file.Name != fm.lastFile.Name
// }

// func (fm *FileMonitor) updateLastFile(file *FileInfo) {
// 	fm.mu.Lock()
// 	defer fm.mu.Unlock()
// 	fm.lastFile = file
// }

func (fm *FileMonitor) printFileInfo(file *FileInfo) {
	fmt.Printf("New file detected: %s (Modified: %s)\n",
		file.FullPath,
		file.ModTime.Format("2006-01-02 15:04:05"))
}

// setupSignalHandler 设置信号处理器
func SetupSignalHandler(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived signal: %v, shutting down...\n", sig)
		cancel()
	}()
}

// 正确做法 - 预编译
func isMatchGood(re *regexp.Regexp, s string) bool {
	return re.MatchString(s)
}

// ProcessData 处理DataFrame数据
func (fi *FileInfo) ProcessData() error {
	dfPtr := fi.df
	df := *dfPtr
	// 1. 去除完全空的行
	// df = dfPtr.Filter(
	// 	dataframe.F{Colname: dfPtr.Names()[0], Comparator: series.Neq, Comparando: ""},
	// )

	// 2. 标准化时间字段
	timeCols := fi.findTimeColumns()

	for _, col := range timeCols {
		df = df.Mutate(
			series.New(df.Col(col).Map(excelToTime), series.String, col),
		)
	}

	// 更新处理后的数据
	fi.df = &df

	return nil
}

// 辅助函数：计算延误时间(分钟)
func calculateDelay(planned, actual series.Series) []interface{} {
	result := make([]interface{}, planned.Len())

	for i := 0; i < planned.Len(); i++ {
		plannedStr := planned.Val(i).(string)
		actualStr := actual.Val(i).(string)

		plannedTime, err1 := time.Parse("2006-01-02 15:04:05", plannedStr)
		actualTime, err2 := time.Parse("2006-01-02 15:04:05", actualStr)

		if err1 != nil || err2 != nil {
			result[i] = 0.0
			continue
		}

		delay := actualTime.Sub(plannedTime).Minutes()
		if delay < 0 {
			delay = 0
		}
		result[i] = delay
	}

	return result
}

// 辅助函数：查找可能是数值类型的列
func (fh *FileInfo) findNumericColumns() []string {
	var numCols []string
	numKeywords := []string{"数量", "金额", "数字", "number", "count", "amount"}

	for _, col := range fh.df.Names() {
		for _, kw := range numKeywords {
			if strings.Contains(col, kw) {
				numCols = append(numCols, col)
				break
			}
		}
	}
	return numCols
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

// 辅助函数：解析时间
func parseTime(v series.Element) series.Element {
	str := v.String()

	// 尝试多种时间格式
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02",
		"01-02-2006 15:04:05",
		"01/02/2006 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			v.Set(t.Format("2006-01-02 15:04:05"))
			return v
		}
	}
	v.Set(str)
	return v
}

// 辅助函数：查找可能是时间类型的列
func (fi *FileInfo) findTimeColumns() []string {
	var timeCols []string
	timeKeywords := []string{"时间", "日期", "date", "time", "COBT"}

	for _, col := range fi.df.Names() {
		for _, kw := range timeKeywords {
			if strings.Contains(col, kw) {
				timeCols = append(timeCols, col)
				break
			}
		}
	}
	return timeCols
}

// 辅助函数：判断DataFrame是否有某列
func (fi *FileInfo) hasColumn(name string) bool {
	for _, n := range fi.df.Names() {
		if n == name {
			return true
		}
	}
	return false
}

func ReadXLSX(filePath, sheetName string) (df dataframe.DataFrame, err error) {

	// 1. 使用tealeg/xlsx打开Excel文件
	xlFile, err := xlsx.OpenFile(filePath)
	if err != nil {
		return dataframe.New(), fmt.Errorf("xlsx open file false: %w", err)
	}

	// 2. 获取第一个工作表
	if len(xlFile.Sheets) == 0 {
		return dataframe.New(), fmt.Errorf("excel文件中没有工作表: %w", err)
	}
	sheet := xlFile.Sheet[sheetName]

	// 3. 转换为Gota DataFrame
	df = convertSheetToDataFrame(sheet)

	return df, err
}

// convertSheetToDataFrame 将xlsx.Sheet转换为dataframe.DataFrame
func convertSheetToDataFrame(sheet *xlsx.Sheet) dataframe.DataFrame {
	if len(sheet.Rows) == 0 {
		return dataframe.New()
	}

	// 获取列名(假设第二行是标题行)
	var headers []string
	for _, cell := range sheet.Rows[1].Cells {
		headers = append(headers, cell.Value)
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
				columns[i] = append(columns[i], cell.Value)
			}
		}
	}

	// 创建Series切片
	seriesList := make([]series.Series, len(headers))
	for i, colName := range headers {
		// 自动推断类型创建Series

		seriesList[i] = series.New(columns[i], series.String, colName)
	}

	return dataframe.New(seriesList...)
}

// updateDelay 计算并添加延误时间列
// func updateDelay(df *dataframe.DataFrame) {
// 	if df == nil {
// 		return
// 	}
// 	plannedCol := "计划起飞时间"
// 	actualCol := "实际起飞时间"
// 	delayCol := "延误时间"

// 	names := df.Names()
// 	hasPlanned, hasActual := false, false
// 	for _, n := range names {
// 		if n == plannedCol {
// 			hasPlanned = true
// 		}
// 		if n == actualCol {
// 			hasActual = true
// 		}
// 	}
// 	if hasPlanned && hasActual {
// 		delay := calculateDelay(df.Col(plannedCol), df.Col(actualCol))
// 		*df = df.Mutate(series.New(delay, series.Float, delayCol))
// 	}
// }

// saveToExcel 将DataFrame保存回Excel文件
func (fh *FileInfo) SaveToExcel(filePath string) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"

	df := fh.df
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

func (fi *FileInfo) addFlightIndex() {
	dfPtr := fi.df
	df := *dfPtr
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
	fi.df = &df
}
