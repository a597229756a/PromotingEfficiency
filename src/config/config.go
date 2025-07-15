package config

import (
	"encoding/json"
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

var (
	instance *Config   // 配置的单例实例
	once     sync.Once // 确保配置只加载一次
)

// LoadConfig 从指定文件夹的JSON文件加载配置
// 参数:
//
//	jsonFolder: JSON文件所在的文件夹路径
//	jsonFile: JSON配置文件名
//
// 返回值:
//
//	*Config: 配置对象的指针
func LoadConfig(jsonFolder, jsonFile string) *Config {
	once.Do(func() {
		// 构建完整的配置文件路径
		configFile := filepath.Join(jsonFolder, jsonFile)
		// 读取配置文件内容
		data, err := os.ReadFile(configFile)
		if err != nil {
			panic("读取配置文件失败: " + err.Error())
		}

		var cfg Config
		// 解析JSON到Config结构体
		if err := json.Unmarshal(data, &cfg); err != nil {
			panic("解析配置失败: " + err.Error())
		}

		instance = &cfg
	})

	return instance
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
