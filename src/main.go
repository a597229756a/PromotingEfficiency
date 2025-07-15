package main

import (
	"PromotingEfficiency/src/config"
	"PromotingEfficiency/src/datasource/email"
	"PromotingEfficiency/src/datasource/file"
	"PromotingEfficiency/src/storage"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 初始化配置
	jsonFolder := "config"
	jsonFile := "config.json"
	cfg := config.LoadConfig(jsonFolder, jsonFile)
	fmt.Println(cfg)

	// 初始化日志系统
	logger, err := storage.NewLogger("app.log")
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	// 启动Web界面显示日志
	go startWebUI(logger)

	// 初始化邮件客户端
	mailClient := email.NewEmailClient(
		cfg.Email.Server,
		cfg.Email.Username,
		cfg.Email.Password,
	)

	// 初始化文件监控
	fileMonitor, err := file.NewFileMonitor(cfg.DataDir)
	if err != nil {
		logger.Fatal("Failed to initialize file monitor:" + err.Error())
	}

	// 启动邮件监控
	go monitorEmails(mailClient, logger, cfg)

	// 启动文件监控
	go monitorFiles(fileMonitor, logger, cfg)

	// 等待退出信号
	waitForShutdown(logger)

}

func monitorEmails(client *email.EmailClient, logger *storage.Logger, cfg *config.Config) {
	handler := email.NewAttachmentHandler(cfg.Email.TargetSubject, cfg.DataDir)

	for {
		emails, err := client.FetchUnreadEmails()
		if err != nil {
			logger.Error("Failed to fetch emails:" + err.Error())
			time.Sleep(5 * time.Minute)
			continue
		}

		for _, e := range emails {
			filePath, err := handler.HandleXLSX(e)
			if err != nil {
				logger.Error("Failed to handle email:" + err.Error())
				continue
			}

			if filePath != "" {
				logger.Info("Downloaded new file: " + filePath)
			}
		}

		time.Sleep(cfg.Email.CheckInterval)
	}
}

func monitorFiles(monitor *file.FileMonitor, logger *storage.Logger, cfg *config.Config) {
	err := monitor.Watch(func(filePath string) {
		df, err := file.ReadXLSXToDataFrame(filePath, "进离港航班")
		if err != nil {
			logger.Error("Failed to read file:" + err.Error())
			return
		}

		// 处理数据...
		logger.Info("Processed updated file: " + filePath)
	})

	if err != nil {
		logger.Error("File monitoring error:" + err.Error())
	}
}

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
