package datapush

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// 常量定义
const (
	APP            = "APP_D1BC57CD7CRJPZXN07FL"
	INFO_FORM      = "FORM-6F4AE8B3490040A29A25A576BCC33438OY6S"
	WARN_FORM      = "FORM-77CB92C91D1E40EABCFEF76C54292AA0T50J"
	SUBS_FORM      = "FORM-61E5F8B49C1647E3BAB11AA2A1A6844AZUHA"
	RUNNING_FORM   = "FORM-62C95C6607254C9DB43679F1265697D1SXSC"
	RUNSUB_FORM    = "FORM-40171AF08DE848738E71BA557B270ED7M5S7"
	DUTY_FORM      = "FORM-J7966ZA162Q3XXKY5GNVYCDM4C5J3SZGHU98L93"
	APP_KEY        = "ding2gb8oi5iswe8mj7q"
	APP_SECRET     = "0yu7UqwlRBtHt5XaI24J1JUMGPJMP_X9t6RY3gN4rLjUePna5iWal2P1fk06HwnC"
	USER_ID        = "0148582454111130975274"
	SYSTEM_TOKEN   = "E0D660C1XV9MW6E1DVNCD8EZS8RV22WOH3TXLP"
	TOKEN_EXPIRE   = 7200 // 2小时
	RETRY_TIMES    = 5
	RETRY_INTERVAL = 2 * time.Second
)

// 全局变量
var (
	accessToken    string
	tokenTimestamp int64
)

// 钉钉 API 响应结构体
type DingTalkResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// Token 响应
type TokenResponse struct {
	DingTalkResponse
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// 上传图片响应
type UploadImageResponse struct {
	DingTalkResponse
	MediaId string `json:"media_id"`
}

// 表单字段映射
var formFieldMappings = map[string]map[string]string{
	"info": {
		"数据日期时间":         "dateField_lxt3od6n",
		"综合效率席短信":        "textareaField_lxt3od6p",
		"航班正常性与延误详情":     "textareaField_lxt3od6r",
		"当前运行概述":         "textareaField_lxt5czt2",
		"CTOT推点航班":       "textField_lxt3od6v",
		"CTOT推点航班图片":     "textField_lxujxpcr",
		"延误1小时以上未起飞航班":   "textField_lxt3od6z",
		"延误1小时以上未起飞航班图片": "textField_lxujxpcs",
	},
	"warn": {
		"航班号": "textField_ly7cbfcg",
		"日期":  "dateField_ly7cbfch",
		"目的地": "textField_lz2cvy1p",
		"机位":  "textField_ly7cbfci",
		"登机门": "textField_lz2cvy1o",
		"地服":  "textField_lyiiagcu",
		"类型":  "textField_ly8a3pkx",
		"描述":  "textField_ly7cbfck",
	},
}

// 获取 AccessToken
func getAccessToken() (string, error) {
	now := time.Now().Unix()

	// 如果 token 未过期，直接返回
	if now-tokenTimestamp < TOKEN_EXPIRE-60 && accessToken != "" {
		return accessToken, nil
	}

	url := fmt.Sprintf("https://oapi.dingtalk.com/gettoken?appkey=%s&appsecret=%s", APP_KEY, APP_SECRET)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("获取 token 请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取 token 响应失败: %v", err)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("解析 token 响应失败: %v", err)
	}

	if tokenResp.ErrCode != 0 {
		return "", fmt.Errorf("获取 token 错误: %s", tokenResp.ErrMsg)
	}

	accessToken = tokenResp.AccessToken
	tokenTimestamp = now
	return accessToken, nil
}

// 上传图片到钉钉
func uploadImage(imagePath string) (string, error) {
	token, err := getAccessToken()
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://oapi.dingtalk.com/media/upload?access_token=%s&type=image", token)

	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("打开图片文件失败: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("media", filepath.Base(imagePath))
	if err != nil {
		return "", fmt.Errorf("创建表单文件失败: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("复制文件内容失败: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("关闭写入器失败: %v", err)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	var uploadResp UploadImageResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if uploadResp.ErrCode != 0 {
		return "", fmt.Errorf("上传图片失败: %s", uploadResp.ErrMsg)
	}

	return uploadResp.MediaId, nil
}

// 批量保存表单数据
func batchSaveFormData(formDataList []map[string]interface{}, formUUID string) error {
	token, err := getAccessToken()
	if err != nil {
		return err
	}

	url := "https://oapi.dingtalk.com/topapi/processinstance/create"

	jsonData, err := json.Marshal(formDataList)
	if err != nil {
		return fmt.Errorf("序列化表单数据失败: %v", err)
	}

	payload := map[string]interface{}{
		"form_uuid":              formUUID,
		"app_type":               APP,
		"system_token":           SYSTEM_TOKEN,
		"user_id":                USER_ID,
		"form_data_json_list":    string(jsonData),
		"no_execute_expression":  false,
		"asynchronous_execution": true,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-acs-dingtalk-access-token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	var result DingTalkResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("保存表单数据失败: %s", result.ErrMsg)
	}

	return nil
}

// 发送钉钉消息
func sendDingMessage(content string, receiverIds []string) error {
	token, err := getAccessToken()
	if err != nil {
		return err
	}

	url := "https://oapi.dingtalk.com/topapi/message/corpconversation/asyncsend_v2"

	payload := map[string]interface{}{
		"agent_id":    "your_agent_id", // 替换为你的应用AgentId
		"userid_list": receiverIds,
		"msg": map[string]interface{}{
			"msgtype": "text",
			"text": map[string]string{
				"content": content,
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-acs-dingtalk-access-token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	var result DingTalkResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("发送消息失败: %s", result.ErrMsg)
	}

	return nil
}

// 重试函数
func retry(fn func() error, times int, interval time.Duration) error {
	var err error
	for i := 0; i < times; i++ {
		if err = fn(); err == nil {
			return nil
		}
		if i < times-1 {
			time.Sleep(interval)
		}
	}
	return fmt.Errorf("重试 %d 次后失败: %v", times, err)
}

// 同步信息数据
func syncInfoData(infoData map[string]interface{}) error {
	formData := make(map[string]interface{})

	for k, v := range infoData {
		if field, ok := formFieldMappings["info"][k]; ok {
			// 处理图片字段
			if field == "textField_lxujxpcr" || field == "textField_lxujxpcs" {
				if imagePath, ok := v.(string); ok && imagePath != "" {
					mediaId, err := uploadImage(imagePath)
					if err != nil {
						log.Printf("上传图片失败: %v", err)
						continue
					}
					formData[field] = mediaId
				}
			} else {
				formData[field] = v
			}
		}
	}

	// 添加更新时间戳
	formData["numberField_lxvgxw12"] = updateCode(infoData["数据日期时间"].(int64))

	return retry(func() error {
		return batchSaveFormData([]map[string]interface{}{formData}, INFO_FORM)
	}, RETRY_TIMES, RETRY_INTERVAL)
}

// 同步告警数据
func syncWarnData(warnData []map[string]interface{}) error {
	var formDataList []map[string]interface{}

	for _, warn := range warnData {
		formData := make(map[string]interface{})
		for k, v := range warn {
			if field, ok := formFieldMappings["warn"][k]; ok {
				formData[field] = v
			}
		}
		formDataList = append(formDataList, formData)
	}

	return retry(func() error {
		return batchSaveFormData(formDataList, WARN_FORM)
	}, RETRY_TIMES, RETRY_INTERVAL)
}

// 更新代码
func updateCode(timestamp int64) int {
	t1 := timestamp % 1800000
	t2 := timestamp % 3600000

	if t2 < 180000 || t2 >= 3480000 {
		return 2
	} else if t1 < 180000 || t1 >= 1680000 {
		return 1
	}
	return 0
}

// func main() {
// 	// 示例使用
// 	infoData := map[string]interface{}{
// 		"数据日期时间":     time.Now().UnixNano() / int64(time.Millisecond),
// 		"综合效率席短信":    "这是测试信息",
// 		"CTOT推点航班图片": "/path/to/image.png",
// 	}

// 	if err := syncInfoData(infoData); err != nil {
// 		log.Fatalf("同步信息数据失败: %v", err)
// 	}

// 	warnData := []map[string]interface{}{
// 		{
// 			"航班号": "CA1234",
// 			"日期":  time.Now().UnixNano() / int64(time.Millisecond),
// 			"类型":  "延误",
// 			"描述":  "航班延误测试",
// 		},
// 	}

// 	if err := syncWarnData(warnData); err != nil {
// 		log.Fatalf("同步告警数据失败: %v", err)
// 	}

// 	log.Println("数据同步完成")
// }
