// data.go
package processor

import (
	"PromotingEfficiency/src/config"
	"PromotingEfficiency/src/utils"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

type DataProcess interface {
	ColCalculation(data *dataframe.DataFrame) error
}

type DelayReasons struct {
	DtDfw   *dataframe.DataFrame
	Dreason dataframe.DataFrame
	Dcfg    *config.DataConfig
	mu      sync.RWMutex
}

func (pdf *DelayReasons) SetDreason(df dataframe.DataFrame) {
	pdf.mu.Lock()
	defer pdf.mu.Unlock()
	pdf.Dreason = df
}

func (pdf *DelayReasons) GetDreason() dataframe.DataFrame {
	pdf.mu.Lock()
	defer pdf.mu.Unlock()
	return pdf.Dreason
}

func (pdf *DelayReasons) SetDtDfw(df *dataframe.DataFrame) {
	pdf.mu.Lock()
	defer pdf.mu.Unlock()
	pdf.DtDfw = df
}

func (pdf *DelayReasons) GetDtDfw() *dataframe.DataFrame {
	pdf.mu.Lock()
	defer pdf.mu.Unlock()
	return pdf.DtDfw
}

func NewDelayReasons(dcfg *config.DataConfig) *DelayReasons {
	return &DelayReasons{
		Dcfg: dcfg,
	}
}

func removeMatchingRowsOptimized(df1, df2 dataframe.DataFrame, keyCol string) dataframe.DataFrame {

	// 1. 提取 df2 的 key 列（如 ID）
	df2Keys := df2.Col(keyCol).Records()

	// 2. 记录 df1 中需要删除的索引
	// var indicesToRemove []int
	// df1Keys := df1.Col(keyCol).Records()[1:]
	// for i, key := range df1Keys {
	// 	for _, df2Key := range df2Keys {
	// 		if key == df2Key {
	// 			indicesToRemove = append(indicesToRemove, i)
	// 			break
	// 		}
	// 	}
	// }
	if len(df2Keys) > 0 {
		df1 = df1.Filter(
			dataframe.F{
				Colname:    keyCol,
				Comparator: series.CompFunc,
				Comparando: func(el series.Element) bool {
					return !Contains(df2Keys, el.String())
				},
			},
		)
	}

	return df1
}
func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func removeRows(df1, df2 dataframe.DataFrame) dataframe.DataFrame {
	// 找出 df1 和 df2 的交集（匹配的行）
	joined := df1.InnerJoin(df2, df1.Names()...)

	// 如果没有交集，直接返回 df1
	if joined.Nrow() == 0 {
		return df1
	}

	// 获取 df1 的所有行 Records
	df1Records := df1.Records()
	joinedRecords := joined.Records()

	// 记录需要删除的索引
	var indicesToRemove []int
	for i, row := range df1Records {
		for _, joinedRow := range joinedRecords {
			match := true
			for colIdx, val := range row {
				if val != joinedRow[colIdx] {
					match = false
					break
				}
			}
			if match {
				indicesToRemove = append(indicesToRemove, i)
				break
			}
		}
	}

	// 删除匹配的行

	if len(indicesToRemove) > 0 {
		df1 = df1.Drop(indicesToRemove)
	}

	return df1
}

func (pdf *DelayReasons) ColCalculation(data *dataframe.DataFrame) error {
	pdf.Dcfg.SetFlightData("atot-stot", "atot-stot")
	if err := utils.SubSeriesTime(data, pdf.Dcfg.GetFlightData("atot"), pdf.Dcfg.GetFlightData("stot"), pdf.Dcfg.GetFlightData("atot-stot")); err != nil {
		return err
	}
	pdf.Dcfg.SetFlightData("sobt-sibt", "sobt-sibt")
	if err := utils.SubSeriesTime(data, pdf.Dcfg.GetFlightData("sobt"), pdf.Dcfg.GetFlightData("sibt"), pdf.Dcfg.GetFlightData("sobt-sibt")); err != nil {
		return err
	}
	pdf.Dcfg.SetFlightData("atot-lastTot", "atot-lastTot")
	if err := utils.SubSeriesTime(data, pdf.Dcfg.GetFlightData("atot"), pdf.Dcfg.GetFlightData("lastTot"), pdf.Dcfg.GetFlightData("atot-lastTot")); err != nil {
		return err
	}

	return nil
}

// updateDelay 是主要的延误处理函数，输入输出都是 DataFrame
func (pdf *DelayReasons) DataProcessFunc(data *dataframe.DataFrame) error {
	cpData := data.Copy()

	// 使用 defer 和 recover 捕获 panic 并记录错误日志
	defer func() {
		if r := recover(); r != nil {
			panic(fmt.Sprintf("自动延误判定失败: %v", r))
		}
	}()

	// 第一步：筛选目标行 - 三级原因 为 NA 的行
	cpData = cpData.Filter(
		dataframe.F{
			Colname:    pdf.Dcfg.GetFlightData("addDelayReason"),
			Comparator: series.CompFunc,
			Comparando: func(el series.Element) bool {
				results := el.IsNA() || el.String() == ""
				return results
			}},
	)

	// 第二步：进一步筛选 - outAtot 填充当前时间后减去 outLastTot 大于 0 的行
	timeDiff := cpData.Filter(
		dataframe.F{
			Colname:    pdf.Dcfg.GetFlightData("atot"),
			Comparator: series.CompFunc,
			Comparando: func(el series.Element) bool {
				if el.IsNA() || el.String() == "" {
					res := time.Now().Format("2006-01-02 15:04:05")
					refV := reflect.ValueOf(res)
					el.Set(refV.Interface())
				}
				return true
			},
		},
	).Copy()

	timeDiff = timeDiff.Filter(dataframe.F{
		Colname:    pdf.Dcfg.GetFlightData("atot-stot"),
		Comparator: series.CompFunc,
		Comparando: func(el series.Element) bool {
			return time.Duration(el.Float()/60) > time.Duration(30)
		},
	})

	// 如果没有符合条件的行，直接返回原始数据
	if timeDiff.Nrow() == 0 {
		pdf.SetDtDfw(data)
		return fmt.Errorf("如果移除后没有剩余行，直接返回 timeDiff Norw == 0")
	}

	// 第三步：移除不符合条件的行
	// 条件：inAldt 为 NA 且 inSldt 不为 NA 且 outSobt - inSibt > 0
	toRemove := timeDiff.Filter(dataframe.F{
		Colname:    pdf.Dcfg.GetFlightData("aldt"),
		Comparator: series.CompFunc,
		Comparando: func(el series.Element) bool {
			results := el.IsNA() || el.String() == ""
			return results
		},
	}).Filter(dataframe.F{
		Colname:    pdf.Dcfg.GetFlightData("sldt"),
		Comparator: series.CompFunc,
		Comparando: func(el series.Element) bool {
			results := el.IsNA() || el.String() == ""
			return results
		},
	}).Filter(dataframe.F{
		Colname:    pdf.Dcfg.GetFlightData("sobt-sibt"),
		Comparator: series.CompFunc,
		Comparando: func(el series.Element) bool {
			return time.Duration(el.Float()) > time.Duration(0)
		},
	})

	// 实际移除行
	if toRemove.Nrow() > 0 {
		timeDiff = removeMatchingRowsOptimized(timeDiff, toRemove, pdf.Dcfg.GetFlightData("flightId"))
	}

	// 如果移除后没有剩余行，直接返回
	if timeDiff.Nrow() == 0 {
		pdf.SetDtDfw(data)
		return fmt.Errorf("如果移除后没有剩余行，直接返回 timeDiff Norw == 0")
	}

	// 根据配置决定如何判定延误原因
	if pdf.Dcfg.GetReasonDelay("均判定为本场天气") == 1 {
		// 全部判定为本场天气
		timeDiff = timeDiff.Mutate(series.New(pdf.Dcfg.Primary["01"], series.String, pdf.Dcfg.GetFlightData("addDelayReason")))
	} else if pdf.Dcfg.ReasonDelay["均判定为本场军事活动"] == 1 {
		// 全部判定为本场军事活动
		timeDiff = timeDiff.Mutate(series.New(pdf.Dcfg.Primary["04"], series.String, pdf.Dcfg.GetFlightData("addDelayReason")))
	} else {
		// 条件1：过站时间严重不足判为公司计划
		if pdf.Dcfg.GetReasonDelay("过站时间严重不足（前序STA晚于后序STD）判为公司计划") == 1 {

			timeDiff = timeDiff.Filter(dataframe.F{
				Colname:    pdf.Dcfg.GetFlightData("sobt-sibt"),
				Comparator: series.CompFunc,
				Comparando: func(el series.Element) bool {
					return time.Duration(el.Float()) <= time.Duration(0)
				},
			})
			timeDiff = timeDiff.Mutate(series.New(pdf.Dcfg.Primary["03"], series.String, pdf.Dcfg.GetFlightData("addDelayReason")))
		}

		// 条件2：根据流控类型判定为外站天气或军事活动
		if pdf.Dcfg.GetReasonDelay("根据流控类型判定为外站天气或军事活动") == 1 {
			filtered := timeDiff.Filter(dataframe.F{
				Colname:    pdf.Dcfg.GetFlightData("addDelayReason"),
				Comparator: series.CompFunc,
				Comparando: func(el series.Element) bool {
					return el.IsNA()
				},
			}).Filter(dataframe.F{
				Colname:    pdf.Dcfg.GetFlightData("tmi"),
				Comparator: series.CompFunc,
				Comparando: func(el series.Element) bool {
					return !el.IsNA()
				},
			})

			if filtered.Ncol() > 0 {
				for i := 0; i < filtered.Ncol(); i++ {
					tmi := filtered.Col("tmi").Elem(i).String()
					if strings.Contains(tmi, "天气") {
						// 判定为外站天气
						timeDiff.Col(pdf.Dcfg.GetFlightData("addDelayReason")).Elem(i).Set("天气")
					} else if strings.Contains(tmi, "其他空域用户") {
						timeDiff.Col(pdf.Dcfg.GetFlightData("addDelayReason")).Elem(i).Set("其他空域用户")
					}

				}
			}
		}
		// 条件3：过站时间不足判为公司计划
		if pdf.Dcfg.GetReasonDelay("过站时间不足（计划过站时间小于最短过站时间）判为公司计划") == 1 {
			filtered := timeDiff.Filter(dataframe.F{
				Colname:    pdf.Dcfg.GetFlightData("addDelayReason"),
				Comparator: series.CompFunc,
				Comparando: func(el series.Element) bool {
					return el.IsNA()
				},
			}).Filter(dataframe.F{
				Colname:    "ttt",
				Comparator: series.Less,
				Comparando: time.Duration(0),
			})

			for _, i := range filtered.Indices() {
				for _, col := range delayReasons {
					timeDiff = timeDiff.Set(delayType(4), series.New([]bool{i == i}, series.Bool, "idx"), col)
				}
			}
		}
	}
	return nil
}

// 	// 剩余延误原因的判定选项
// 	remainingOptions := []string{
// 		"剩余延误原因判定为本场天气",
// 		"剩余延误原因判定为本场军事活动",
// 		"剩余延误原因判定为外站天气",
// 		"剩余延误原因判定为外站军事活动",
// 	}

// 	// 检查剩余选项，使用第一个匹配的配置
// 	for idx, option := range remainingOptions {
// 		if dp.DELAYBY[option] {
// 			filtered := timeDiff.Filter(dataframe.F{
// 				Colname:    "rstDelayReason",
// 				Comparator: series.IsNA,
// 			})
// 			for _, col := range delayReasons {
// 				timeDiff = timeDiff.Set(delayType(idx), filtered.Indices(), col)
// 			}
// 			break
// 		}
// 	}
// }

// // 准备更新数据
// var updates []map[string]interface{}
// // 筛选出已判定延误原因的行
// filtered := timeDiff.Filter(dataframe.F{
// 	Colname:    "rstDelayReason",
// 	Comparator: series.NotIsNA,
// })

// // 构建更新数据结构
// for _, i := range filtered.Indices() {
// 	update := make(map[string]interface{})
// 	flightNo := filtered.Col("outFlightNo").Elem(i).String()
// 	rstDelayReason := filtered.Col("rstDelayReason").Elem(i).String()

// 	// 尝试获取 delayGuid，如果有效则使用 guid 模式
// 	delayGuid := filtered.Col("delayGuid").Elem(i)
// 	if delayGuid.IsValid() && !delayGuid.IsNA() {
// 		guid, err := delayGuid.Int()
// 		if err == nil {
// 			update["flightNo"] = flightNo
// 			update["guid"] = guid
// 			update["rstDelayReason"] = rstDelayReason
// 		}
// 	} else {
// 		// 否则使用 flightDate 和 sobt 模式
// 		update["flightNo"] = flightNo
// 		flightDate := filtered.Col("outFlightDate").Elem(i).String()
// 		if len(flightDate) > 19 {
// 			flightDate = flightDate[:19] // 截取前19个字符
// 		}
// 		update["flightDate"] = flightDate

// 		sobt := filtered.Col("outSobt").Elem(i).String()
// 		if len(sobt) > 19 {
// 			sobt = sobt[:19] // 截取前19个字符
// 		}
// 		update["sobt"] = sobt

// 		update["rstDelayReason"] = rstDelayReason
// 	}

// 	// 如果有附加延误原因，添加到更新数据
// 	addDelayReason := filtered.Col("addDelayReason").Elem(i)
// 	if addDelayReason.IsValid() && !addDelayReason.IsNA() && addDelayReason.String() != "" {
// 		update["addDelayReason"] = addDelayReason.String()
// 	}

// 	updates = append(updates, update)
// }

// // 将判定结果合并回原始数据
// for _, col := range delayReasons {
// 	data = data.Join(timeDiff.Select([]string{col}), dataframe.JoinType(0))
// }

// // 如果有需要更新的数据，启动 goroutine 异步处理
// if len(updates) > 0 {
// 	go func() {
// 		// 确保 goroutine 中的 panic 被捕获
// 		defer func() {
// 			if r := recover(); r != nil {
// 				log.Printf("Error in update_delay_data: %v", r)
// 			}
// 		}()
// 		dp.updateDelayData(updates)
// 	}()
// }

// 	return nil
// }
