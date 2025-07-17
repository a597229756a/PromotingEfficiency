package main

import (
	"PromotingEfficiency/src/config"
	"PromotingEfficiency/src/datasource/email"
	"PromotingEfficiency/src/storage"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron"
)

func main() {
	jsonFolder := "./config"
	jsonFile := "config.json"
	dataJsonFile := "dataconfig.json"
	cfg, dcfg, err := config.LoadConfig(jsonFolder, jsonFile, dataJsonFile)

	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	fmt.Println(dcfg)

	if err != nil {
		fmt.Println(err)
	}

	// 初始化日志系统
	logger, err := storage.NewLogger("app.log")
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	// 邮箱地址，用户名和密码
	emailClient := email.NewEmailClient(
		cfg.Email.Server,
		cfg.Email.Username,
		cfg.Email.Password)

	// 连接到邮箱
	handler := email.NewXLSXAttachmentHandler(cfg.Email.TargetSubject, cfg.DataDir)

	var dfw *email.DataFrameWrapper = &email.DataFrameWrapper{}
	// 设置定时任务
	c := cron.New()

	// 使用配置中的检查间隔而不是硬编码的1分钟
	interval := time.Duration(cfg.Email.CheckInterval).String() // 例如 "5m0s"
	cronSpec := fmt.Sprintf("@every %s", interval)

	// 添加定时任务
	err = c.AddFunc(cronSpec, func() {
		logger.Info(fmt.Sprintf("开始定时检查(间隔: %v)...", cronSpec))

		t1 := time.Now()
		// 查询email是否有更新
		newEmail, err := email.CheckAndProcessEmails(emailClient, logger)
		if err != nil {
			logger.Error("检查处理邮件失败: " + err.Error()) // 使用Error级别记录错误
		}

		// 将附件另存为xlsx
		go func() {
			if err := handler.Handle(newEmail, logger); err != nil {
				logger.Error(fmt.Sprintf("处理邮件失败(UID:%d): %v", newEmail.UID, err))
			}
		}()

		// 附件转dataframe错误
		bytes := newEmail.Attachments[0].Content
		if err := dfw.ReadXLSX(bytes, "进离港航班"); err != nil {
			logger.Error(err.Error())
		}
		t2 := time.Since(t1)
		logger.Info(fmt.Sprintf("数据处理时间：%v\n", t2))

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

// func monitorEmails(client *email.EmailClient, logger *storage.Logger, cfg *config.Config) {
// 	handler := email.NewAttachmentHandler(cfg.Email.TargetSubject, cfg.DataDir)

// 	for {
// 		emails, err := client.FetchUnreadEmails()
// 		if err != nil {
// 			logger.Error("Failed to fetch emails:" + err.Error())
// 			time.Sleep(5 * time.Minute)
// 			continue
// 		}

// 		for _, e := range emails {
// 			filePath, err := handler.HandleXLSX(e)
// 			if err != nil {
// 				logger.Error("Failed to handle email:" + err.Error())
// 				continue
// 			}

// 			if filePath != "" {
// 				logger.Info("Downloaded new file: " + filePath)
// 			}
// 		}

// 		time.Sleep(cfg.Email.CheckInterval)
// 	}
// }

// func monitorFiles(monitor *file.FileMonitor, logger *storage.Logger, cfg *config.Config) {
// 	err := monitor.Watch(func(filePath string) {
// 		df, err := file.ReadXLSXToDataFrame(filePath, "进离港航班")
// 		if err != nil {
// 			logger.Error("Failed to read file:" + err.Error())
// 			return
// 		}

// 		// 处理数据...
// 		logger.Info("Processed updated file: " + filePath)
// 	})

// 	if err != nil {
// 		logger.Error("File monitoring error:" + err.Error())
// 	}
// }

// startWebUI 启动一个简单的Web界面来显示实时日志
// 参数:
//
//	logger: 日志记录器实例，用于订阅日志消息
func startWebUI(logger *storage.Logger) {
	// 注册/logs路由的处理函数
	http.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		// 设置响应头
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Transfer-Encoding", "chunked")

		// 创建日志订阅通道
		logChan := logger.Subscribe()

		// 无限循环，持续接收日志消息
		for {
			select {
			case msg := <-logChan:
				// 将日志消息写入HTTP响应
				_, err := fmt.Fprintln(w, msg)
				if err != nil {
					// 如果写入失败(如客户端断开连接)，则退出循环
					return
				}
				// 刷新响应缓冲区，确保消息立即发送到客户端
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			case <-r.Context().Done():
				// 如果客户端断开连接，则退出循环
				return
			}
		}
	})

	// 可以在这里添加更多路由或启动服务器的代码
	http.ListenAndServe(":8080", nil)
}

func waitForShutdown(logger *storage.Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logger.Info("Received signal: " + sig.String() + ", shutting down...")
	logger.Close()
	os.Exit(0)

}
