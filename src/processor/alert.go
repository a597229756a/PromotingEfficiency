package processor

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"strings"
// 	"time"

// 	"github.com/go-gota/gota/dataframe"
// 	"github.com/go-gota/gota/series"
// )

// const (
// 	TypeCode = "your_typecode" // 替换为实际的航班类型代码
// )

// type AlertResponse struct {
// 	Data struct {
// 		Data1 int `json:"data1"`
// 		Data2 int `json:"data2"`
// 		Data5 int `json:"data5"`
// 		Data6 int `json:"data6"`
// 	} `json:"data"`
// }

// type AlertSystem struct {
// 	URLs   map[string]string
// 	HEADER map[string]string
// 	// 其他需要的字段
// }

// func (a *AlertSystem) datetimeNow() time.Time {
// 	return time.Now()
// }

// func (a *AlertSystem) getSession() *http.Client {
// 	return &http.Client{
// 		Timeout: time.Second * 10,
// 	}
// }

// func (a *AlertSystem) updateLog(msg string, level string) {
// 	log.Printf("[%s] %s\n", level, msg)
// }

// func (a *AlertSystem) getData(dataType string, flightTypeCode string) dataframe.DataFrame {
// 	// 实现获取数据的逻辑，返回 Gota DataFrame
// 	// 这里需要根据您的实际数据源实现
// 	return dataframe.DataFrame{}
// }

// func (a *AlertSystem) checkLargeScaleFlightDelay() map[string]interface{} {
// 	datetimeNow := a.datetimeNow()
// 	alert := []int{-1}
// 	alertMap := map[int][]string{
// 		-1: {"无", ""},
// 		0:  {"预警标准"},
// 		1:  {"黄色响应"},
// 		2:  {"橙色响应"},
// 		3:  {"红色响应"},
// 	}
// 	alerts := make(map[string]interface{})

// 	// 尝试从API获取数据
// 	session := a.getSession()
// 	resp, err := session.Post(a.URLs["大面积航延"], "application/json", nil)
// 	if err != nil {
// 		a.updateLog("大面积航延接口数据获取失败，采用计算值", "warn")
// 		return a.calculateAlerts(datetimeNow, alert, alertMap, alerts)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		a.updateLog(fmt.Sprintf("响应异常代码%d", resp.StatusCode), "warn")
// 		return a.calculateAlerts(datetimeNow, alert, alertMap, alerts)
// 	}

// 	var alertResp AlertResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&alertResp); err != nil {
// 		a.updateLog("解析响应数据失败", "warn")
// 		return a.calculateAlerts(datetimeNow, alert, alertMap, alerts)
// 	}

// 	alerts["预警条件"] = alertResp.Data.Data1
// 	alerts["响应条件一"] = alertResp.Data.Data2
// 	alerts["响应条件二"] = alertResp.Data.Data5
// 	alerts["响应条件三"] = alertResp.Data.Data6

// 	return a.evaluateAlerts(datetimeNow, alert, alertMap, alerts)
// }

// func (a *AlertSystem) calculateAlerts(datetimeNow time.Time, alert []int, alertMap map[int][]string, alerts map[string]interface{}) map[string]interface{} {
// 	// 大面积航班延误预警
// 	flightData := a.getData("航班", TypeCode)
// 	flightData = flightData.Filter(
// 		dataframe.F{
// 			Colname:    "outAtot",
// 			Comparator: series.IsNA,
// 		},
// 	).Filter(
// 		dataframe.F{
// 			Colname:    "outEstripStatus",
// 			Comparator: series.Neq,
// 			Comparando: "DEP",
// 		},
// 	)

// 	// 计算预警条件
// 	ctotCondition := flightData.Col("outCtot").Map(func(s series.Series) series.Series {
// 		return s.Map(func(e interface{}) interface{} {
// 			t, ok := e.(time.Time)
// 			if !ok || t.Before(datetimeNow) {
// 				return datetimeNow
// 			}
// 			return t
// 		})
// 	}).Sub(flightData.Col("outSobt")).Map(func(s series.Series) series.Series {
// 		return s.Map(func(e interface{}) interface{} {
// 			d, ok := e.(time.Duration)
// 			if !ok {
// 				return false
// 			}
// 			return d >= time.Hour*3/2 // 1.5小时
// 		})
// 	})
// 	alerts["预警条件"] = flightData.Filter(
// 		dataframe.F{
// 			Colname:    "outCtot",
// 			Comparator: series.CompFunc,
// 			Comparando: func(idx int, s series.Series) bool {
// 				val, err := s.Elem(idx).Bool()
// 				return err == nil && val
// 			},
// 			Col: ctotCondition,
// 		},
// 	).Nrow()

// 	// 计算响应条件一
// 	stotCondition := flightData.Col("outStot").Map(func(s series.Series) series.Series {
// 		return s.Map(func(e interface{}) interface{} {
// 			t, ok := e.(time.Time)
// 			if !ok {
// 				return false
// 			}
// 			return datetimeNow.Sub(t) > time.Hour
// 		})
// 	})
// 	alerts["响应条件一"] = flightData.Filter(
// 		dataframe.F{
// 			Colname:    "outStot",
// 			Comparator: series.CompFunc,
// 			Comparando: func(idx int, s series.Series) bool {
// 				val, err := s.Elem(idx).Bool()
// 				return err == nil && val
// 			},
// 			Col: stotCondition,
// 		},
// 	).Nrow()

// 	// 计算响应条件二、三
// 	departureData := a.getData("执行_离港", TypeCode)
// 	canceledFlights := departuredata.Filter(
// 		dataframe.F{
// 			Colname:    "operationStatusCode",
// 			Comparator: series.Eq,
// 			Comparando: "CNCL",
// 		},
// 	)

// 	// 响应条件二
// 	condition2 := canceledFlights.Filter(
// 		dataframe.F{
// 			Colname:    "sobt",
// 			Comparator: series.Gte,
// 			Comparando: datetimeNow,
// 		},
// 	).Filter(
// 		dataframe.F{
// 			Colname:    "sobt",
// 			Comparator: series.Lt,
// 			Comparando: datetimeNow.Add(time.Hour * 2),
// 		},
// 	)
// 	alerts["响应条件二"] = condition2.Nrow()

// 	// 响应条件三
// 	condition3 := canceledFlights.Filter(
// 		dataframe.F{
// 			Colname:    "sobt",
// 			Comparator: series.Gte,
// 			Comparando: datetimeNow,
// 		},
// 	).Filter(
// 		dataframe.F{
// 			Colname:    "sobt",
// 			Comparator: series.Lt,
// 			Comparando: datetimeNow.Add(time.Hour),
// 		},
// 	)
// 	arrivalData := a.getData("执行_进港", TypeCode)
// 	condition3Arrival := arrivalData.Filter(
// 		dataframe.F{
// 			Colname:    "operationStatusCode",
// 			Comparator: series.Neq,
// 			Comparando: "CNCL",
// 		},
// 	).Filter(
// 		dataframe.F{
// 			Colname:    "eldt",
// 			Comparator: series.Gte,
// 			Comparando: datetimeNow,
// 		},
// 	).Filter(
// 		dataframe.F{
// 			Colname:    "eldt",
// 			Comparator: series.Lt,
// 			Comparando: datetimeNow.Add(time.Hour),
// 		},
// 	)
// 	alerts["响应条件三"] = condition3.Nrow() + condition3Arrival.Nrow()

// 	return a.evaluateAlerts(datetimeNow, alert, alertMap, alerts)
// }

// func (a *AlertSystem) evaluateAlerts(datetimeNow time.Time, alert []int, alertMap map[int][]string, alerts map[string]interface{}) map[string]interface{} {
// 	// 评估预警条件
// 	if val, ok := alerts["预警条件"].(int); ok && val >= 10 {
// 		alertMap[0] = append(alertMap[0], fmt.Sprintf("离港航班CTOT≥STD+1.5小时达10架次及以上（当前%d架次）", val))
// 		alert = append(alert, 0)
// 	}

// 	// 评估响应条件一
// 	if val, ok := alerts["响应条件一"].(int); ok {
// 		thresholds := []struct {
// 			value int
// 			level int
// 		}{
// 			{50, 3},
// 			{40, 2},
// 			{25, 1},
// 		}
// 		for _, t := range thresholds {
// 			if val >= t.value {
// 				alertMap[t.level] = append(alertMap[t.level],
// 					fmt.Sprintf("离港航班不正常时长超1小时以上且未起飞客运航班达%d架次（按起飞延误，当前%d架次）", t.value, val))
// 				alert = append(alert, t.level)
// 				break
// 			}
// 		}
// 	}

// 	// 评估响应条件二
// 	if val, ok := alerts["响应条件二"].(int); ok {
// 		thresholds := []struct {
// 			value int
// 			level int
// 		}{
// 			{30, 3},
// 			{25, 2},
// 			{15, 1},
// 		}
// 		for _, t := range thresholds {
// 			if val >= t.value {
// 				alertMap[t.level] = append(alertMap[t.level],
// 					fmt.Sprintf("未来2小时内出港客运航班航司决策临时取消达%d架次（当前%d架次）", t.value, val))
// 				alert = append(alert, t.level)
// 				break
// 			}
// 		}
// 	}

// 	// 评估响应条件三
// 	if val, ok := alerts["响应条件三"].(int); ok {
// 		hour := datetimeNow.Hour()
// 		if hour >= 22 || hour < 6 {
// 			thresholds := []struct {
// 				value int
// 				level int
// 			}{
// 				{75, 3},
// 				{70, 2},
// 				{60, 1},
// 			}
// 			for _, t := range thresholds {
// 				if val >= t.value {
// 					alertMap[t.level] = append(alertMap[t.level],
// 						fmt.Sprintf("22时至次日6时任一小时进港航班和离港取消航班合计达%d架次（当前%d架次）", t.value, val))
// 					alert = append(alert, t.level)
// 					break
// 				}
// 			}
// 		} else {
// 			delete(alerts, "响应条件三")
// 		}
// 	}

// 	// 确定最高警报级别
// 	maxAlert := -1
// 	for _, level := range alert {
// 		if level > maxAlert {
// 			maxAlert = level
// 		}
// 	}

// 	// 准备返回结果
// 	result := make(map[string]interface{})
// 	result["大面积航延"] = alertMap[maxAlert][0]

// 	// 合并启动标准
// 	var standards []string
// 	if len(alertMap[maxAlert]) > 1 {
// 		standards = alertMap[maxAlert][1:]
// 	}
// 	result["启动标准"] = strings.Join(standards, "；")

// 	return result
// }
