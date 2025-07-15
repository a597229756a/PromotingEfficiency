package email

import (
	"PromotingEfficiency/src/config"
	"PromotingEfficiency/src/storage"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/robfig/cron"
)

func TestSendEmail(t *testing.T) {
	jsonFolder := "../../config"
	jsonFile := "config.json"
	cfg := config.LoadConfig(jsonFolder, jsonFile)

	SendEmail(cfg)
}

func TestNewEmailClient(t *testing.T) {
	jsonFolder := "../../config"
	jsonFile := "config.json"
	cfg := config.LoadConfig(jsonFolder, jsonFile)

	// 初始化日志系统
	logger, err := storage.NewLogger("app.log")
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	// 邮箱地址，用户名和密码
	emailClient := NewEmailClient(
		cfg.Email.Server,
		cfg.Email.Username,
		cfg.Email.Password)

	// 连接到邮箱
	handler := NewXLSXAttachmentHandler(cfg.Email.TargetSubject, cfg.DataDir)

	// 设置定时任务
	c := cron.New()

	// 使用配置中的检查间隔而不是硬编码的1分钟
	interval := time.Duration(cfg.Email.CheckInterval).String() // 例如 "5m0s"
	cronSpec := fmt.Sprintf("@every %s", interval)

	// 添加定时任务
	err = c.AddFunc(cronSpec, func() {
		logger.Info(fmt.Sprintf("开始定时检查(间隔: %v)...", cronSpec))

		if err := checkAndProcessEmails(emailClient, handler, logger); err != nil {
			logger.Error("检查处理邮件失败: " + err.Error()) // 使用Error级别记录错误
		}
	})

	if err != nil {
		logger.Error("创建定时任务失败: " + err.Error())
		return // 重要错误应该终止程序
	}

	// 启动定时任务
	c.Start()
	defer c.Stop() // 接收 Stop() 返回的 context

	logger.Info(fmt.Sprintf("邮件监控服务已启动(检查间隔: %v)，按Ctrl+C退出", interval))
	select {}

}
