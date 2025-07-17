package processor

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"strings"
// 	"time"

// 	"github.com/go-gota/gota/dataframe"
// )

// type Flight struct {
// 	GUID               string
// 	FlightNo           string
// 	RouteCn            string
// 	AirportRegionCn    string
// 	GateNo             string
// 	BoardingStatus     string
// 	STD                time.Time
// 	CTOT               time.Time
// 	CTOTOffset         time.Duration
// 	TTT                time.Duration
// 	InboundFlightTime  string
// 	Notes              string
// 	CTOTHistory        string
// 	TrafficControlInfo string
// }

// type CTOTResult struct {
// 	Summary      string
// 	Details      dataframe.DataFrame
// 	ExcelPath    string
// 	ImagePath    string
// 	WarningNotes []string
// }

// func (s *Service) fetchFlightData(flightType string) (dataframe.DataFrame, error) {
// 	// 从数据库或API获取基础航班数据
// 	query := fmt.Sprintf("SELECT * FROM flights WHERE outFlightTypeCode IN (%s)", flightType)
// 	df, err := s.db.Query(query)
// 	if err != nil {
// 		return dataframe.DataFrame{}, fmt.Errorf("查询航班数据失败: %w", err)
// 	}

// 	// 过滤已有实际起飞时间或未来航班
// 	now := time.Now()
// 	filtered := df.Filter(
// 		dataframe.F{Colname: "outAtot", Comparator: series.IsNA},
// 	).Or(
// 		dataframe.F{Colname: "outAtot", Comparator: series.Greater, Comparando: now},
// 	)

// 	return filtered, nil
// }

// func calculateCTOTOffset(df dataframe.DataFrame, delayColumn string) dataframe.DataFrame {
// 	// 计算CTOT与STOT/SOBT的差值
// 	ctot := df.Col("outCtot").Records()
// 	base := df.Col(delayColumn).Records()

// 	offsets := make([]string, len(ctot))
// 	for i := range ctot {
// 		ctotTime, _ := time.Parse(time.RFC3339, ctot[i])
// 		baseTime, _ := time.Parse(time.RFC3339, base[i])
// 		offset := ctotTime.Sub(baseTime)
// 		offsets[i] = offset.String()
// 	}

// 	return df.Mutate(series.New(offsets, series.String, "ctotOffset"))
// }

// func (s *Service) analyzeCTOTHistory(df dataframe.DataFrame) dataframe.DataFrame {
// 	history := s.historyDB.Query("SELECT guid, ctot_history FROM ctot_history")

// 	joined := df.Join(history, "left", "guid")

// 	// 应用历史分析逻辑
// 	// ...

// 	return joined
// }

// func generateFlightStatusNotes(df dataframe.DataFrame, now time.Time) dataframe.DataFrame {
// 	statusMap := map[string]string{
// 		"outPushTime":                  "滑行中",
// 		"moniJob.tract_D.actBeginTime": "已推出",
// 		"outAcct":                      "已关舱",
// 		"outAebt":                      "已登结",
// 		"outAsbt":                      "登机中",
// 	}

// 	notes := make([]string, df.Nrow())
// 	for i := 0; i < df.Nrow(); i++ {
// 		var status string
// 		for col, desc := range statusMap {
// 			colTime, _ := time.Parse(time.RFC3339, df.Col(col).Elem(i).String())
// 			if !colTime.IsZero() && colTime.Before(now) {
// 				status += desc + "\n"
// 			}
// 		}
// 		notes[i] = strings.TrimSpace(status)
// 	}

// 	return df.Mutate(series.New(notes, series.String, "outStatus"))
// }

// func (s *Service) CalculateCTOT(minutes int, outputPath string) (*CTOTResult, error) {
// 	// 1. 获取数据
// 	df, err := s.fetchFlightData(s.ctotType)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 2. 计算CTOT偏移量
// 	delayColumn := "outStot"
// 	if s.delayType == "outStot" {
// 		delayColumn = "outStot"
// 	}
// 	df = calculateCTOTOffset(df, delayColumn)

// 	// 3. 过滤分钟数
// 	minOffset := time.Duration(minutes) * time.Minute
// 	df = df.Filter(
// 		dataframe.F{Colname: "ctotOffset", Comparator: series.Greater, Comparando: minOffset},
// 	)

// 	// 4. 分析历史数据
// 	df = s.analyzeCTOTHistory(df)

// 	// 5. 生成状态备注
// 	df = generateFlightStatusNotes(df, time.Now())

// 	// 6. 生成结果
// 	result, err := s.generateResult(df, minutes, outputPath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return result, nil
// }

// func (s *Service) generateResult(df dataframe.DataFrame, minutes int, path string) (*CTOTResult, error) {
// 	result := &CTOTResult{
// 		WarningNotes: s.collectWarnings(),
// 	}

// 	if df.Nrow() == 0 {
// 		result.Summary = "无CTOT推点航班。"
// 		return result, nil
// 	}

// 	// 生成统计信息
// 	stats := calculateStatistics(df)
// 	result.Summary = formatSummary(stats, minutes)

// 	// 导出Excel
// 	excelPath, err := exportToExcel(df, path)
// 	if err != nil {
// 		return nil, fmt.Errorf("导出Excel失败: %w", err)
// 	}
// 	result.ExcelPath = excelPath

// 	// 生成图片
// 	imgPath, err := exportToImage(excelPath)
// 	if err != nil {
// 		result.WarningNotes = append(result.WarningNotes, "图片生成失败")
// 	} else {
// 		result.ImagePath = imgPath
// 	}

// 	return result, nil
// }

// func (s *Service) CalculateCTOTConcurrently(minutes int, outputPath string) (*CTOTResult, error) {
// 	var (
// 		df  dataframe.DataFrame
// 		err error
// 	)

// 	// 使用errgroup管理并发任务
// 	g, ctx := errgroup.WithContext(context.Background())

// 	// 并发获取数据
// 	g.Go(func() error {
// 		data, e := s.fetchFlightData(s.ctotType)
// 		if e != nil {
// 			return e
// 		}
// 		df = data
// 		return nil
// 	})

// 	// 并发获取历史数据
// 	var history dataframe.DataFrame
// 	g.Go(func() error {
// 		data, e := s.historyDB.Query("SELECT guid, ctot_history FROM ctot_history")
// 		if e != nil {
// 			return e
// 		}
// 		history = data
// 		return nil
// 	})

// 	if err := g.Wait(); err != nil {
// 		return nil, err
// 	}

// 	// 合并数据
// 	df = df.Join(history, "left", "guid")

// 	// 继续处理...
// }

// func exportToExcel(df dataframe.DataFrame, path string) (string, error) {
// 	if path == "" {
// 		var err error
// 		path, err = promptSavePath()
// 		if err != nil {
// 			return "", fmt.Errorf("获取保存路径失败: %w", err)
// 		}
// 	}

// 	if err := os.MkdirAll(path, 0755); err != nil {
// 		return "", fmt.Errorf("创建目录失败: %w", err)
// 	}

// 	file := filepath.Join(path, generateFilename())
// 	if err := df.WriteExcel(file); err != nil {
// 		return "", fmt.Errorf("写入Excel文件失败: %w", err)
// 	}

// 	return file, nil
// }
