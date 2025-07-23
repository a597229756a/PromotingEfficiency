package processor

// import (
// 	"fmt"
// 	"time"

// 	"github.com/go-gota/gota/dataframe"
// 	"github.com/go-gota/gota/series"
// )

// // 分析航班数据
// func analyzeFlightData(dfDeparture, dfArrival, dfCancellation dataframe.DataFrame, today time.Time, now time.Time) (initial, clearance, takeoff, landing string, outCancel, inCancel int) {
// 	// 初始化结果
// 	initial = ""
// 	clearance = ""
// 	takeoff = ""
// 	landing = ""
// 	outCancel = 0
// 	inCancel = 0

// 	// 1. 始发航班正常性分析
// 	initialData := dfDeparture.Filter(
// 		dataframe.F{Colname: "outIsinitial", Comparator: series.Eq, Comparando: true},
// 	)
// 	if initialData.Nrow() > 0 {
// 		// 尚未离港的始发航班
// 		initialDeparting := initialData.FilterAggregation(
// 			dataframe.And,
// 			dataframe.F{Colname: "outAtot", Comparator: series.IsNA, Comparando: nil},
// 			dataframe.F{Colname: "outAtot", Comparator: series.Greater, Comparando: now},
// 		)
// 		initialDepartingDelayed := initialDeparting.Filter(
// 			dataframe.F{Colname: "outStot", Comparator: series.Less, Comparando: now},
// 		).Nrow()

// 		// 移除尚未离港的航班
// 		initialDepartedData := initialData.Drop(initialDeparting.Indices())

// 		// 正常离港的始发航班
// 		initialDeparted := initialDepartedData.Filter(
// 			dataframe.F{Colname: "outAtot", Comparator: series.LessEq, Comparando: initialDepartedData.Col("outStot")},
// 		).Nrow()

// 		// 延误离港的始发航班
// 		initialDelayed := initialDepartedData.Nrow() - initialDeparted

// 		// 计算始发正常率
// 		total := initialDeparted + initialDepartingDelayed + initialDelayed
// 		if total > 0 {
// 			rate := float64(initialDeparted) / float64(total)
// 			if rate == 1 {
// 				initial = "100%"
// 			} else {
// 				initial = fmt.Sprintf("%.2f%%", rate*100)
// 			}
// 		}
// 	}

// 	// 2. 放行正常性分析
// 	departing := dfDeparture.FilterAggregation(
// 		dataframe.Or,
// 		dataframe.F{Colname: "outAtot", Comparator: series.IsNA, Comparando: nil},
// 		dataframe.F{Colname: "outAtot", Comparator: series.Greater, Comparando: now},
// 	)

// 	// 放行延误的航班
// 	clearanceDepartingDelayed := departing.FilterAggregation(
// 		dataframe.And,
// 		dataframe.Filter{
// 			Colname:    "inAldt",
// 			Comparator: series.NotIsNA,
// 			Comparando: nil,
// 		},
// 		dataframe.Filter{
// 			Colname:    "inSldt",
// 			Comparator: series.IsNA,
// 			Comparando: nil,
// 		},
// 		dataframe.Filter{
// 			Colname:    "outLastTot",
// 			Comparator: series.Less,
// 			Comparando: now,
// 		},
// 	).Nrow()

// 	// 起飞延误的航班
// 	takeoffDepartingDelayed := departing.Filter(
// 		dataframe.F{Colname: "outStot", Comparator: series.Less, Comparando: now},
// 	).Nrow()

// 	// 已离港航班
// 	departedData := dfDeparture.Drop(departing.Indices())

// 	if departedData.Nrow() > 0 {
// 		// 正常放行的航班
// 		clearanceDeparted := departedData.Filter(
// 			dataframe.F{Colname: "outAtot", Comparator: series.LessEq, Comparando: departedData.Col("outLastTot")},
// 		).Nrow()

// 		// 放行延误的航班
// 		clearanceDelayed := departedData.FilterAggregation(
// 			dataframe.Or,
// 			dataframe.F{Colname: "inAldt", Comparator: series.NotIsNA, Comparando: nil},
// 			dataframe.F{Colname: "inSldt", Comparator: series.IsNA, Comparando: nil},
// 		).Nrow() - clearanceDeparted

// 		// 计算放行正常率
// 		total := clearanceDeparted + clearanceDepartingDelayed + clearanceDelayed
// 		if total > 0 {
// 			rate := float64(clearanceDeparted) / float64(total)
// 			if rate == 1 {
// 				clearance = "100%"
// 			} else {
// 				clearance = fmt.Sprintf("%.2f%%", rate*100)
// 			}
// 		}

// 		// 3. 起飞正常性分析
// 		takeoffDeparted := departedData.Filter(
// 			dataframe.F{Colname: "outAtot", Comparator: series.LessEq, Comparando: departedData.Col("outStot")},
// 		).Nrow()

// 		takeoffDelayed := departedData.Nrow() - takeoffDeparted

// 		// 计算起飞正常率
// 		total = takeoffDeparted + takeoffDepartingDelayed + takeoffDelayed
// 		if total > 0 {
// 			rate := float64(takeoffDeparted) / float64(total)
// 			if rate == 1 {
// 				takeoff = "100%"
// 			} else {
// 				takeoff = fmt.Sprintf("%.2f%%", rate*100)
// 			}
// 		}
// 	}

// 	// 4. 进港正常性分析
// 	if dfArrival.Nrow() > 0 {
// 		// 尚未到达的航班
// 		arriving := dfArrival.FilterAggregation(
// 			dataframe.Or,
// 			dataframe.F{Colname: "aldt", Comparator: series.IsNA, Comparando: nil},
// 			dataframe.F{Colname: "aldt", Comparator: series.Greater, Comparando: now},
// 		)

// 		// 进港延误的航班
// 		landingArrivingDelayed := arriving.Filter(
// 			dataframe.F{
// 				Colname:    "sibt",
// 				Comparator: series.Less,
// 				Comparando: now.Add(-10 * time.Minute),
// 			},
// 		).Nrow()

// 		// 已到达航班
// 		arrivedData := dfArrival.Drop(arriving.Indices())

// 		if arrivedData.Nrow() > 0 {
// 			// 正常到达的航班
// 			landingArrived := arrivedData.Filter(
// 				dataframe.F{
// 					Colname:    "aldt",
// 					Comparator: series.LessEq,
// 					Comparando: arrivedData.Col("sibt").Add(10 * time.Minute),
// 				},
// 			).Nrow()

// 			// 到达延误的航班
// 			landingDelayed := arrivedData.Nrow() - landingArrived

// 			// 计算进港正常率
// 			total := landingArrived + landingArrivingDelayed + landingDelayed
// 			if total > 0 {
// 				rate := float64(landingArrived) / float64(total)
// 				if rate == 1 {
// 					landing = "100%"
// 				} else {
// 					landing = fmt.Sprintf("%.2f%%", rate*100)
// 				}
// 			}
// 		}
// 	}

// 	// 5. 取消航班统计
// 	if dfCancellation.Nrow() > 0 {
// 		outCancel = dfCancellation.Filter(
// 			dataframe.F{Colname: "operationStatusCode", Comparator: series.Eq, Comparando: "CNCL"},
// 		).Nrow()
// 	}

// 	// 进港取消航班统计
// 	if dfArrival.Nrow() > 0 {
// 		inCancel = dfArrival.Filter(
// 			dataframe.F{Colname: "operationStatusCode", Comparator: series.Eq, Comparando: "CNCL"},
// 		).Nrow()
// 	}

// 	return initial, clearance, takeoff, landing, outCancel, inCancel
// }
