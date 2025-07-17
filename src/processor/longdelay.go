package processor

// import (
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"strings"
// 	"time"

// 	"github.com/go-gota/gota/dataframe"
// 	"github.com/go-gota/gota/series"
// )

// type DelayAnalysis struct {
// 	DelayColumn      string
// 	ExcludeNoArrival bool
// 	StatusMap        map[string]string
// 	InfoPath         string
// }

// type DelayResult struct {
// 	Summary   string
// 	ExcelPath string
// 	ImagePath string
// 	Warnings  []string
// }

// func (da *DelayAnalysis) LongDelay(hours float64, outputPath string) (*DelayResult, error) {
// 	result := &DelayResult{}

// 	// 准备警告信息
// 	warnings := da.prepareWarnings(hours)

// 	// 确保输出目录存在
// 	if err := os.MkdirAll(outputPath, 0755); err != nil {
// 		return nil, fmt.Errorf("创建目录失败: %w", err)
// 	}

// 	// 获取当前时间
// 	now := time.Now()

// 	// 1. 获取航班数据
// 	df, err := da.fetchFlightData()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 2. 过滤数据
// 	filtered := da.filterFlights(df, now, hours)
// 	if filtered.Nrow() == 0 {
// 		result.Summary = fmt.Sprintf("截至%02d:%02d，无延误%.1f小时以上未起飞航班。",
// 			now.Hour(), now.Minute(), hours)
// 		return result, nil
// 	}

// 	// 3. 计算延误时间
// 	withTimes := da.calculateDelayTimes(filtered, now)

// 	// 4. 准备输出数据
// 	outputDF := da.prepareOutputData(withTimes, now)

// 	// 5. 生成统计信息
// 	stats := da.generateStatistics(outputDF, hours, now)

// 	// 6. 导出Excel
// 	excelPath, err := da.exportToExcel(outputDF, outputPath, hours)
// 	if err != nil {
// 		return nil, err
// 	}
// 	result.ExcelPath = excelPath

// 	// 7. 生成图片 (示例，实际需要调用相应库)
// 	imgPath, err := da.exportToImage(excelPath)
// 	if err != nil {
// 		warnings = append(warnings, "图片生成失败")
// 	} else {
// 		result.ImagePath = imgPath
// 	}

// 	// 组装结果
// 	result.Summary = stats
// 	result.Warnings = warnings

// 	return result, nil
// }

// // 准备警告信息
// func (da *DelayAnalysis) prepareWarnings(hours float64) []string {
// 	var warnings []string

// 	if da.DelayColumn == "outStot" {
// 		warnings = append(warnings, "按起飞延误")
// 	}

// 	if da.ExcludeNoArrival {
// 		warnings = append(warnings, "不含前序未落地航班")
// 	}

// 	return warnings
// }

// // 获取航班数据
// func (da *DelayAnalysis) fetchFlightData() (dataframe.DataFrame, error) {
// 	// 这里应该是从数据库或文件获取原始数据
// 	// 示例使用假数据
// 	df := dataframe.LoadRecords(
// 		[][]string{
// 			{"outFlightNo", "outRouteCn", "inAldt", "outGateNo", "outSobt", "outStot", "outCtot"},
// 			{"CA1234", "北京-上海", "2023-01-01 12:00", "A1", "2023-01-01 13:00", "2023-01-01 13:30", "2023-01-01 14:00"},
// 			{"MU5678", "上海-广州", "2023-01-01 11:00", "B2", "2023-01-01 12:00", "2023-01-01 12:15", "2023-01-01 13:00"},
// 		},
// 	)

// 	return df, nil
// }

// // 过滤航班数据
// func (da *DelayAnalysis) filterFlights(df dataframe.DataFrame, now time.Time, hours float64) dataframe.DataFrame {
// 	// 过滤未起飞或起飞时间在未来
// 	filtered := df.Filter(
// 		dataframe.F{Colname: "outAtot", Comparator: series.IsNA},
// 	).Or(
// 		dataframe.F{Colname: "outAtot", Comparator: series.Greater, Comparando: now},
// 	)

// 	// 过滤延误航班
// 	filtered = filtered.Filter(
// 		dataframe.F{Colname: da.DelayColumn, Comparator: series.Less, Comparando: now},
// 	)

// 	// 如果需要排除前序未落地
// 	if da.ExcludeNoArrival {
// 		filtered = filtered.Filter(
// 			dataframe.F{Colname: "inAldt", Comparator: series.NotIsNA},
// 		)
// 	}

// 	// 过滤超过指定延误时间的
// 	minDelay := time.Duration(hours * float64(time.Hour))
// 	filtered = filtered.Filter(
// 		dataframe.F{
// 			Colname: da.DelayColumn,
// 			Comparator: func(s series.Series) series.Series {
// 				delays := make([]bool, s.Len())
// 				for i := 0; i < s.Len(); i++ {
// 					t, err := time.Parse("2006-01-02 15:04", s.Elem(i).String())
// 					if err != nil {
// 						delays[i] = false
// 						continue
// 					}
// 					delays[i] = now.Sub(t) >= minDelay
// 				}
// 				return series.Bools(delays)
// 			},
// 		},
// 	)

// 	return filtered
// }

// // 计算延误时间
// func (da *DelayAnalysis) calculateDelayTimes(df dataframe.DataFrame, now time.Time) dataframe.DataFrame {
// 	// 计算等待时间
// 	waiting := make([]string, df.Nrow())
// 	delayed := make([]string, df.Nrow())
// 	toCtot := make([]string, df.Nrow())

// 	for i := 0; i < df.Nrow(); i++ {
// 		// 计算机上等待时间
// 		acct, _ := time.Parse("2006-01-02 15:04", df.Col("outAcct").Elem(i).String())
// 		waiting[i] = now.Sub(acct).String()

// 		// 计算已延误时间
// 		delayTime, _ := time.Parse("2006-01-02 15:04", df.Col(da.DelayColumn).Elem(i).String())
// 		delayed[i] = now.Sub(delayTime).String()

// 		// 计算距CTOT时间
// 		ctot, _ := time.Parse("2006-01-02 15:04", df.Col("outCtot").Elem(i).String())
// 		toCtot[i] = ctot.Sub(now).String()
// 	}

// 	return df.Mutate(
// 		series.New(waiting, series.String, "waiting"),
// 		series.New(delayed, series.String, "delayed"),
// 		series.New(toCtot, series.String, "toCtot"),
// 	)
// }

// // 准备输出数据
// func (da *DelayAnalysis) prepareOutputData(df dataframe.DataFrame, now time.Time) dataframe.DataFrame {
// 	// 添加状态信息
// 	status := make([]string, df.Nrow())
// 	for i := 0; i < df.Nrow(); i++ {
// 		for col, desc := range da.StatusMap {
// 			colTime, _ := time.Parse("2006-01-02 15:04", df.Col(col).Elem(i).String())
// 			if !colTime.IsZero() && colTime.Before(now) {
// 				status[i] = desc
// 				break
// 			}
// 		}
// 	}

// 	// 格式化时间显示
// 	inAldt := make([]string, df.Nrow())
// 	for i := 0; i < df.Nrow(); i++ {
// 		t, err := time.Parse("2006-01-02 15:04", df.Col("inAldt").Elem(i).String())
// 		if err == nil {
// 			inAldt[i] = t.Format("01-02 15:04")
// 		} else {
// 			inAldt[i] = "前站未起"
// 		}
// 	}

// 	return df.Mutate(
// 		series.New(status, series.String, "outStatus"),
// 		series.New(inAldt, series.String, "inAldt"),
// 	)
// }

// // 生成统计信息
// func (da *DelayAnalysis) generateStatistics(df dataframe.DataFrame, hours float64, now time.Time) string {
// 	hourStr := fmt.Sprintf("%.1f小时以上", hours)
// 	if hours == float64(int(hours)) {
// 		hourStr = fmt.Sprintf("%d小时以上", int(hours))
// 	}

// 	// 按区域分组统计
// 	grouped := df.GroupBy("outAirportRegionCn")
// 	groups := grouped.GetGroups()

// 	var regionCounts []string
// 	for region, group := range groups {
// 		regionCounts = append(regionCounts, fmt.Sprintf("%s%d架次", region, group.Nrow()))
// 	}

// 	// 构建统计信息
// 	stats := fmt.Sprintf("截至%02d:%02d，延误%s未起飞航班%d架次（%s）",
// 		now.Hour(), now.Minute(), hourStr, df.Nrow(), strings.Join(regionCounts, "，"))

// 	// 按延误时间段统计
// 	delayGroups := make(map[string]int)
// 	for i := 0; i < df.Nrow(); i++ {
// 		delayStr := df.Col("delayed").Elem(i).String()
// 		d, _ := time.ParseDuration(delayStr)
// 		hours := int(d.Hours())

// 		key := fmt.Sprintf("%d-%d小时", hours, hours+1)
// 		delayGroups[key]++
// 	}

// 	for key, count := range delayGroups {
// 		stats += fmt.Sprintf("，延误%s%d架次", key, count)
// 	}

// 	return stats + "。"
// }

// // 导出到Excel
// func (da *DelayAnalysis) exportToExcel(df dataframe.DataFrame, path string, hours float64) (string, error) {
// 	filename := fmt.Sprintf("延误%.1f小时未起飞航班_%s.xlsx", hours, time.Now().Format("20060102_150405"))
// 	filePath := filepath.Join(path, filename)

// 	// 选择需要导出的列
// 	output := df.Select([]string{
// 		"outFlightNo", "outRouteCn", "inAldt", "outGateNo",
// 		"outSobt", da.DelayColumn, "outCtot", "delayed",
// 		"toCtot", "waiting", "outStatus", "outAirportRegionCn",
// 	})

// 	// 重命名列
// 	output = output.Rename("航班号", "outFlightNo").
// 		Rename("下站", "outRouteCn").
// 		Rename("前序航班落地时间", "inAldt").
// 		Rename("登机门", "outGateNo").
// 		Rename("STD", "outSobt").
// 		Rename("起延", da.DelayColumn).
// 		Rename("CTOT", "outCtot").
// 		Rename("已延误", "delayed").
// 		Rename("距CTOT", "toCtot").
// 		Rename("机上等待", "waiting").
// 		Rename("状态", "outStatus").
// 		Rename("区域", "outAirportRegionCn")

// 	// 写入Excel文件
// 	f, err := os.Create(filePath)
// 	if err != nil {
// 		return "", fmt.Errorf("创建文件失败: %w", err)
// 	}
// 	defer f.Close()

// 	if err := output.WriteExcel(f); err != nil {
// 		return "", fmt.Errorf("写入Excel失败: %w", err)
// 	}

// 	return filePath, nil
// }

// // 导出图片 (示例)
// func (da *DelayAnalysis) exportToImage(excelPath string) (string, error) {
// 	// 实际实现需要调用相关库将Excel转为图片
// 	return excelPath[:len(excelPath)-5] + ".png", nil
// }
