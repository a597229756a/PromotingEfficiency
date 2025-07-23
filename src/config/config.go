package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Config 结构体定义了应用程序的配置结构
type Config struct {
	Email struct {
		Server        string   `json:"server"`         // 邮件服务器地址
		Username      string   `json:"username"`       // 邮箱用户名
		Password      string   `json:"password"`       // 邮箱密码
		TargetSubject string   `json:"target_subject"` // 需要匹配的邮件主题
		CheckInterval Duration `json:"check_interval"` // 检查新邮件的间隔时间
	} `json:"email"`

	DataDir    string `json:"data_dir"` // 应用程序数据存储目录
	SheetName  string `json:"sheet_name"`
	LogName    string `json:"log_name"`
	LogMaxSize string `json:"log_max_size"`
	SendEmail  struct {
		Server        string `json:"server"`         // 邮件服务器地址
		Username      string `json:"username"`       // 邮箱用户名
		Password      string `json:"password"`       // 邮箱密码
		TargetSubject string `json:"target_subject"` // 需要匹配的邮件主题
		Attachment    string `json:"attachment"`     // 检查新邮件的间隔时间
	} `json:"send_email"`
}

type DataConfig struct {
	AircraftSize map[string]int    `json:"aircraftsize"`
	Primary      map[string]string `json:"primary"`
	ReasonDelay  map[string]int    `json:"reasondelay"`
	FlightData   map[string]string `json:"flightData"`
}

var (
	once               sync.Once
	instance           *Config
	dataConfigInstance *DataConfig
	mu                 sync.RWMutex
)

func LoadConfig(jsonFolder, jsonFile, dataJsonFile string) (*Config, *DataConfig, error) {
	var err error
	once.Do(func() {
		instance, dataConfigInstance, err = loadConfigs(jsonFolder, jsonFile, dataJsonFile)
	})
	return instance, dataConfigInstance, err
}

func loadConfigs(jsonFolder, jsonFile, dataJsonFile string) (*Config, *DataConfig, error) {
	configFile := filepath.Join(jsonFolder, jsonFile)
	dataConfigFile := filepath.Join(jsonFolder, dataJsonFile)

	configData, err := readFile(configFile)
	if err != nil {
		return nil, nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	dataConfigData, err := readFile(dataConfigFile)
	if err != nil {
		return nil, nil, fmt.Errorf("读取数据配置文件失败: %w", err)
	}

	cfgChan := make(chan *Config, 1)
	dcfgChan := make(chan *DataConfig, 1)
	errChan := make(chan error, 2)

	go parseConfig(configData, cfgChan, errChan)
	go parseDataConfig(dataConfigData, dcfgChan, errChan)

	cfg, dcfg, err := waitForResults(cfgChan, dcfgChan, errChan)
	if err != nil {
		return nil, nil, err
	}

	return cfg, dcfg, nil
}

func readFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法读取文件 %s: %w", filePath, err)
	}
	return data, nil
}

func parseConfig(data []byte, resultChan chan<- *Config, errChan chan<- error) {
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		errChan <- fmt.Errorf("解析Config失败: %w", err)
		return
	}
	resultChan <- &cfg
}

func parseDataConfig(data []byte, resultChan chan<- *DataConfig, errChan chan<- error) {
	var dcfg DataConfig
	if err := json.Unmarshal(data, &dcfg); err != nil {
		errChan <- fmt.Errorf("解析DataConfig失败: %w", err)
		return
	}
	resultChan <- &dcfg
}

func waitForResults(
	cfgChan <-chan *Config,
	dcfgChan <-chan *DataConfig,
	errChan <-chan error,
) (*Config, *DataConfig, error) {
	var (
		cfg    *Config
		dcfg   *DataConfig
		errors []error
	)

	for i := 0; i < 2; i++ {
		select {
		case c := <-cfgChan:
			cfg = c
			fmt.Println("Config 配置文件加载完毕")
		case d := <-dcfgChan:
			dcfg = d
			fmt.Println("DataConfig 配置文件加载完毕")
		case err := <-errChan:
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return nil, nil, combineErrors(errors)
	}

	if cfg == nil || dcfg == nil {
		return nil, nil, fmt.Errorf("部分配置未加载成功")
	}

	return cfg, dcfg, nil
}

func combineErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	// 使用固定格式字符串
	msg := "配置加载遇到多个错误:"
	for _, err := range errs {
		msg = fmt.Sprintf("%s\n- %v", msg, err)
	}
	return fmt.Errorf("%s", msg)
}

// Duration 是time.Duration的自定义包装类型
// 用于支持JSON序列化和反序列化
type Duration time.Duration

// UnmarshalJSON 实现json.Unmarshaler接口
// 用于从JSON字符串解析Duration
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

// MarshalJSON 实现json.Marshaler接口
// 用于将Duration序列化为JSON字符串
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (dc *DataConfig) GetFlightData(colName string) string {
	mu.RLock()
	defer mu.RUnlock()
	return dc.FlightData[colName]
}

func (dc *DataConfig) SetFlightData(colName, value string) {
	mu.Lock()
	defer mu.Unlock()
	dc.FlightData[colName] = value
}

func (dc *DataConfig) GetReasonDelay(colName string) int {
	mu.RLock()
	defer mu.RUnlock()
	return dc.ReasonDelay[colName]
}

func (dc *DataConfig) SetReasonDelay(colName string, value int) {
	mu.Lock()
	defer mu.Unlock()
	dc.ReasonDelay[colName] = value
}
