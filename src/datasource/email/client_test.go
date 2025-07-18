package email

import (
	"PromotingEfficiency/src/config"
	"PromotingEfficiency/src/storage"
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/robfig/cron"
)

func TestSendEmail(t *testing.T) {
	jsonFolder := "../../config"
	jsonFile := "config.json"
	dataJsonFile := "dataconfig.json"
	cfg, dcfg, err := config.LoadConfig(jsonFolder, jsonFile, dataJsonFile)
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	fmt.Println(dcfg)

	SendEmail(cfg)
	fmt.Println(time.Now())
}

func TestNewEmailClient(t *testing.T) {
	jsonFolder := "../../config"
	jsonFile := "config.json"
	dataJsonFile := "dataconfig.json"
	cfg, dcfg, err := config.LoadConfig(jsonFolder, jsonFile, dataJsonFile)
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	fmt.Println(dcfg)

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

	var dfw *DataFrameWrapper = &DataFrameWrapper{}
	// 设置定时任务
	c := cron.New()

	// 使用配置中的检查间隔而不是硬编码的1分钟
	interval := time.Duration(cfg.Email.CheckInterval).String() // 例如 "5m0s"
	cronSpec := fmt.Sprintf("@every %s", interval)
		
	// 添加定时任务
	err = c.AddFunc(cronSpec, func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Email.CheckInterval))
		defer cancel()

		logger.Info(fmt.Sprintf("开始定时检查(间隔: %v)...", cronSpec))

		// 查询email是否有更新
		newEmail, err := CheckAndProcessEmails(ctx, emailClient, handler, logger)

		if err != nil {
			logger.Error("检查处理邮件失败: " + err.Error()) // 使用Error级别记录错误
		} else if !handler.isProcessed(newEmail.UID) { // 如果邮件没有处理过

			t1 := time.Now()
			// 处理数据
			bytes := newEmail.Attachments[0].Content
			if err := dfw.ReadXLSX(bytes, "进离港航班"); err != nil {
				logger.Error(err.Error())
			}

			worker(ctx, handler, newEmail, logger, dfw)
			t2 := time.Since(t1)
			logger.Info(fmt.Sprintf("数据处理使用时间: %v...", t2))

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

func worker(ctx context.Context, handler *XLSXAttachmentHandler, newEmail *Email, logger *storage.Logger, dfw *DataFrameWrapper) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := handler.Handle(newEmail, logger); err != nil {
			logger.Error(fmt.Sprintf("处理邮件失败(UID:%d): %v", newEmail.UID, err))
		}
	}()
	go func() {
		defer wg.Done()
		// pdf := processor.NewDelayReasons(dcfg)
		// if err := pdf.DataProcessFunc(&dfw.df); err != nil {
		// 	logger.Error(err.Error())
		// }
	}()

	// select {
	// case <-ctx.Done():
	// 	fmt.Printf("Worker exiting (context cancelled)\n")
	// 	return
	// }
	wg.Wait()
}
