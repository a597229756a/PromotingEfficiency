package processor

// import (
// 	"fmt"
// 	"strings"
// 	"time"

// 	"github.com/go-gota/gota/dataframe"
// 	"github.com/go-gota/gota/series"
// )

// type FlightStatus struct {
// 	StatusMap    map[string]string
// 	RunwayDirMap map[string]string
// 	ApronGateMap map[string]string
// 	CtotType     []string
// }

// type FlightOverview struct {
// 	fs *FlightStatus
// }

// func (fo *FlightOverview) GenerateOverview(isYesterday bool, outputPath string) (string, error) {
// 	now := time.Now()
// 	today := fo.getDate(isYesterday)
// 	tomorrow := today.AddDate(0, 0, 1)

// 	// 获取数据
// 	arrivals, departures, flights, passengers, err := fo.getFlightData()
// 	if err != nil {
// 		return "", err
// 	}

// 	// 过滤取消的航班
// 	departures = fo.filterCancelledFlights(departures)
// 	arrivals = fo.filterCancelledFlights(arrivals)

// 	// 生成基础统计信息
// 	msg := fo.generateBasicStats(arrivals, departures, today, tomorrow, now, isYesterday)

// 	// 生成未执行航班信息
// 	msg += fo.generateUnflownStats(flights, today, tomorrow, now, isYesterday)

// 	// 生成未登机旅客信息
// 	if !isYesterday && passengers.Nrow() > 0 {
// 		msg += fo.generatePassengerStats(passengers)
// 	}

// 	// 生成准备出港航班信息
// 	msg += fo.generateReadyForDepartureStats(departures)

// 	// 生成CTOT推点信息
// 	msg += fo.generateCtotStats(flights, now)

// 	// 生成预计进港航班信息
// 	msg += fo.generateEstimatedArrivals(arrivals, now)

// 	return msg, nil
// }

// func (fo *FlightOverview) getDate(isYesterday bool) time.Time {
// 	now := time.Now()
// 	if isYesterday {
// 		return now.AddDate(0, 0, -1)
// 	}
// 	return now
// }

// func (fo *FlightOverview) getFlightData() (arrivals, departures, flights, passengers dataframe.DataFrame, err error) {
// 	// 这里应该是从数据库或文件获取原始数据
// 	// 示例使用假数据
// 	arrivals = dataframe.LoadRecords([][]string{
// 		{"flightNo", "sibt", "aldt", "flightTypeCode"},
// 		{"CA1234", "2023-01-01 12:00", "2023-01-01 12:30", "J"},
// 		{"MU5678", "2023-01-01 13:00", "", "J"},
// 	})

// 	departures = dataframe.LoadRecords([][]string{
// 		{"flightNo", "sobt", "atot", "flightTypeCode", "operationStatusCode", "asbt", "runwayTypeCode"},
// 		{"CA1235", "2023-01-01 14:00", "2023-01-01 14:30", "J", "SCH", "2023-01-01 13:30", "01"},
// 		{"MU5679", "2023-01-01 15:00", "", "J", "SCH", "2023-01-01 14:30", "02"},
// 	})

// 	flights = dataframe.LoadRecords([][]string{
// 		{"outFlightNo", "outSobt", "outAtot", "outCtot", "outEstripStatus", "outAirportRegionCn"},
// 		{"CA1235", "2023-01-01 14:00", "2023-01-01 14:30", "2023-01-01 14:20", "DEP", "A区"},
// 		{"MU5679", "2023-01-01 15:00", "", "2023-01-01 15:30", "", "B区"},
// 	})

// 	passengers = dataframe.LoadRecords([][]string{
// 		{"flightNo", "acctCabin", "gateNo", "unBoardNum"},
// 		{"CA1235", "Y", "A1", "50"},
// 		{"MU5679", "", "B2", "30"},
// 	})

// 	return arrivals, departures, flights, passengers, nil
// }

// func (fo *FlightOverview) filterCancelledFlights(df dataframe.DataFrame) dataframe.DataFrame {
// 	return df.Filter(
// 		dataframe.F{Colname: "operationStatusCode", Comparator: series.Neq, Comparando: "CNCL"},
// 	)
// }

// func (fo *FlightOverview) generateBasicStats(arrivals, departures dataframe.DataFrame,
// 	today, tomorrow, now time.Time, isYesterday bool) string {

// 	// 计划航班统计
// 	plannedDep := departures.Filter(
// 		dataframe.F{Colname: "sobt", Comparator: series.GreaterEq, Comparando: today},
// 	).Filter(
// 		dataframe.F{Colname: "sobt", Comparator: series.Less, Comparando: tomorrow},
// 	).Filter(
// 		dataframe.F{Colname: "flightTypeCode", Comparator: series.Neq, Comparando: "F/H"},
// 	).Filter(
// 		dataframe.F{Colname: "flightTypeCode", Comparator: series.Neq, Comparando: "Q/B"},
// 	)

// 	plannedArr := arrivals.Filter(
// 		dataframe.F{Colname: "sibt", Comparator: series.GreaterEq, Comparando: today},
// 	).Filter(
// 		dataframe.F{Colname: "sibt", Comparator: series.Less, Comparando: tomorrow},
// 	).Filter(
// 		dataframe.F{Colname: "flightTypeCode", Comparator: series.Neq, Comparando: "F/H"},
// 	).Filter(
// 		dataframe.F{Colname: "flightTypeCode", Comparator: series.Neq, Comparando: "Q/B"},
// 	)

// 	// 已执行航班统计
// 	flownDep := plannedDep.Filter(
// 		dataframe.F{Colname: "atot", Comparator: series.NotIsNA},
// 	)

// 	flownArr := plannedArr.Filter(
// 		dataframe.F{Colname: "aldt", Comparator: series.NotIsNA},
// 	)

// 	// 实际执行航班统计
// 	actualDep := departures.Filter(
// 		dataframe.F{Colname: "sobt", Comparator: series.GreaterEq, Comparando: today},
// 	).Filter(
// 		dataframe.F{Colname: "atot", Comparator: series.Less, Comparando: now},
// 	)

// 	actualArr := arrivals.Filter(
// 		dataframe.F{Colname: "sibt", Comparator: series.GreaterEq, Comparando: today},
// 	).Filter(
// 		dataframe.F{Colname: "aldt", Comparator: series.Less, Comparando: now},
// 	)

// 	// 构建消息
// 	dayDesc := "昨日"
// 	if !isYesterday {
// 		dayDesc = ""
// 	}

// 	return fmt.Sprintf("截至%02d:%02d，%s计划执行航班%d架次（离港%d架次，进港%d架次），其中已执行%d架次（离港%d架次，进港%d架次）；\n",
// 		now.Hour(), now.Minute(),
// 		dayDesc,
// 		plannedDep.Nrow()+plannedArr.Nrow(),
// 		plannedDep.Nrow(),
// 		plannedArr.Nrow(),
// 		flownDep.Nrow()+flownArr.Nrow(),
// 		flownDep.Nrow(),
// 		flownArr.Nrow(),
// 	) + fmt.Sprintf("实际执行%d架次（离港%d架次，进港%d架次）；\n",
// 		actualDep.Nrow()+actualArr.Nrow(),
// 		actualDep.Nrow(),
// 		actualArr.Nrow(),
// 	)
// }

// func (fo *FlightOverview) generateUnflownStats(flights dataframe.DataFrame,
// 	today, tomorrow, now time.Time, isYesterday bool) string {

// 	// 未起飞航班
// 	unflown := flights.Filter(
// 		dataframe.F{Colname: "outAtot", Comparator: series.IsNA},
// 	).Filter(
// 		dataframe.F{Colname: "outSobt", Comparator: series.GreaterEq, Comparando: today},
// 	).Filter(
// 		dataframe.F{Colname: "outSobt", Comparator: series.Less, Comparando: tomorrow},
// 	)

// 	if unflown.Nrow() == 0 {
// 		if isYesterday {
// 			return "昨日离港航班均已起飞；"
// 		}
// 		return fmt.Sprintf("今日计划时间在%02d:%02d前的离港航班均已起飞；", now.Hour(), now.Minute())
// 	}

// 	// 添加状态信息
// 	statusCol := make([]string, unflown.Nrow())
// 	for i := 0; i < unflown.Nrow(); i++ {
// 		for k, v := range fo.fs.StatusMap {
// 			if !unflown.Col(k).Elem(i).IsNA() {
// 				statusCol[i] = v
// 				break
// 			}
// 		}
// 	}
// 	unflown = unflown.Mutate(series.New(statusCol, series.String, "status"))

// 	// 统计各状态数量
// 	statusCounts := make(map[string]int)
// 	for _, s := range statusCol {
// 		statusCounts[s]++
// 	}

// 	// 构建消息
// 	var statusParts []string
// 	for status, count := range statusCounts {
// 		statusParts = append(statusParts, fmt.Sprintf("%s%d架次", status, count))
// 	}

// 	return fmt.Sprintf("今日计划时间在%02d:%02d前的离港航班%d架次未执行（%s）；",
// 		now.Hour(), now.Minute(),
// 		unflown.Nrow(),
// 		strings.Join(statusParts, "，"),
// 	)
// }

// func (fo *FlightOverview) generatePassengerStats(passengers dataframe.DataFrame) string {
// 	// 未登机旅客
// 	unboarded := passengers.Filter(
// 		dataframe.F{Colname: "acctCabin", Comparator: series.IsNA},
// 	)

// 	// 按区域分组
// 	unboarded = unboarded.Mutate(
// 		series.New(
// 			unboarded.Col("gateNo").Map(func(s series.Series) series.Series {
// 				if s.IsNA() {
// 					return series.Strings("未发布登机门航班旅客")
// 				}
// 				return series.Strings(fo.fs.ApronGateMap[s.Elem(0).String()])
// 			}),
// 			series.String, "parkRegion",
// 		),
// 	)

// 	// 确保未登机人数非负
// 	unboarded = unboarded.Mutate(
// 		series.New(
// 			unboarded.Col("unBoardNum").Map(func(s series.Series) series.Series {
// 				n := s.Elem(0).Int()
// 				if n < 0 {
// 					return series.Ints(0)
// 				}
// 				return series.Ints(n)
// 			}),
// 			series.Int, "unBoardNum",
// 		),
// 	)

// 	total := unboarded.Col("unBoardNum").Sum()
// 	if total == 0 {
// 		return "\n今日所有未执行的离港航班中，已过检未登机旅客约0人；"
// 	}

// 	// 按区域统计
// 	var regionParts []string
// 	regions := []string{"A指廊", "B指廊", "C指廊", "D指廊", "E指廊",
// 		"国内西远机位", "国内东远机位", "国际远机位", "未发布登机门航班旅客"}

// 	for _, region := range regions {
// 		sum := unboarded.Filter(
// 			dataframe.F{Colname: "parkRegion", Comparator: series.Eq, Comparando: region},
// 		).Col("unBoardNum").Sum()

// 		if sum > 0 {
// 			regionParts = append(regionParts, fmt.Sprintf("%s约%d人", region, sum))
// 		}
// 	}

// 	return fmt.Sprintf("\n今日所有未执行的离港航班中，已过检未登机旅客约%d人（%s）；",
// 		total, strings.Join(regionParts, "，"))
// }

// func (fo *FlightOverview) generateReadyForDepartureStats(departures dataframe.DataFrame) string {
// 	// 已登机未起飞航班
// 	ready := departures.Filter(
// 		dataframe.F{Colname: "atot", Comparator: series.IsNA},
// 	).Filter(
// 		dataframe.F{Colname: "asbt", Comparator: series.NotIsNA},
// 	)

// 	if ready.Nrow() == 0 {
// 		return "\n无已登机准备出港航班"
// 	}

// 	// 添加跑道方向信息
// 	ready = ready.Mutate(
// 		series.New(
// 			ready.Col("runwayTypeCode").Map(func(s series.Series) series.Series {
// 				return series.Strings(fo.fs.RunwayDirMap[s.Elem(0).String()])
// 			}),
// 			series.String, "runwayDir",
// 		),
// 	)

// 	// 按跑道方向分组
// 	groups := ready.GroupBy("runwayDir").GetGroups()

// 	var runwayParts []string
// 	for dir, group := range groups {
// 		taxiing := group.Filter(
// 			dataframe.F{Colname: "pushTime", Comparator: series.NotIsNA},
// 		).Nrow()

// 		runwayParts = append(runwayParts,
// 			fmt.Sprintf("%s向%d架次，其中%d架次滑行中", dir, group.Nrow(), taxiing))
// 	}

// 	return fmt.Sprintf("\n已登机准备出港航班%d架次（%s）；",
// 		ready.Nrow(), strings.Join(runwayParts, "；"))
// }

// func (fo *FlightOverview) generateCtotStats(flights dataframe.DataFrame, now time.Time) string {
// 	// CTOT推点航班
// 	ctotFlights := flights.Filter(
// 		dataframe.F{Colname: "outCtot", Comparator: series.NotIsNA},
// 	).Filter(
// 		dataframe.F{Colname: "outAtot", Comparator: series.IsNA},
// 	).Filter(
// 		dataframe.F{Colname: "outEstripStatus", Comparator: series.Neq, Comparando: "DEP"},
// 	)

// 	// 计算推点时长
// 	ctotFlights = ctotFlights.Mutate(
// 		series.New(
// 			ctotFlights.Col("outCtot").Map(func(s series.Series) series.Series {
// 				ctot, _ := time.Parse("2006-01-02 15:04", s.Elem(0).String())
// 				delay, _ := time.Parse("2006-01-02 15:04", flights.Col("outLastTot").Elem(0).String())
// 				return series.Strings(ctot.Sub(delay).String())
// 			}),
// 			series.String, "ctotOffset",
// 		),
// 	)

// 	// 过滤正推点
// 	ctotFlights = ctotFlights.Filter(
// 		dataframe.F{Colname: "ctotOffset", Comparator: series.Greater, Comparando: "0s"},
// 	)

// 	if ctotFlights.Nrow() == 0 {
// 		return "无CTOT推点航班。"
// 	}

// 	// 按区域分组
// 	regionGroups := ctotFlights.GroupBy("outAirportRegionCn").GetGroups()

// 	var regionParts []string
// 	for region, group := range regionGroups {
// 		regionParts = append(regionParts, fmt.Sprintf("%s%d架次", region, group.Nrow()))
// 	}

// 	// 计算平均推点时长
// 	avgOffset := ctotFlights.Col("ctotOffset").Mean()

// 	// 推点1小时以上
// 	longDelay := ctotFlights.Filter(
// 		dataframe.F{Colname: "ctotOffset", Comparator: series.Greater, Comparando: "1h"},
// 	).Nrow()

// 	avgDesc := ""
// 	if ctotFlights.Nrow() > 1 {
// 		avgDesc = "平均"
// 	}

// 	return fmt.Sprintf("CTOT推点航班共计%d架次（%s），%s推点时长%s，推点一小时以上%d架次。",
// 		ctotFlights.Nrow(),
// 		strings.Join(regionParts, "，"),
// 		avgDesc,
// 		formatDuration(avgOffset),
// 		longDelay,
// 	)
// }

// func (fo *FlightOverview) generateEstimatedArrivals(arrivals dataframe.DataFrame, now time.Time) string {
// 	// 未来一小时预计进港
// 	estimated := arrivals.Filter(
// 		dataframe.F{Colname: "eldt", Comparator: series.GreaterEq, Comparando: now},
// 	).Filter(
// 		dataframe.F{Colname: "eldt", Comparator: series.Less, Comparando: now.Add(time.Hour)},
// 	)

// 	return fmt.Sprintf("未来一小时预计进港航班%d架次。", estimated.Nrow())
// }

// func formatDuration(seconds float64) string {
// 	d := time.Duration(seconds) * time.Second
// 	h := d / time.Hour
// 	d -= h * time.Hour
// 	m := d / time.Minute

// 	if h > 0 {
// 		return fmt.Sprintf("%d小时%02d分钟", h, m)
// 	}
// 	return fmt.Sprintf("%d分钟", m)
// }
