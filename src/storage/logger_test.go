package storage

import (
	"PromotingEfficiency/src/config"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	code := m.Run()

	jsonFolder := "../config"
	jsonFile := "config.json"
	cfg := config.LoadConfig(jsonFolder, jsonFile)
	fmt.Println(cfg)

	// 初始化logger记录器
	logger, err := NewLogger(cfg.LogName)
	if err != nil {
		logger.Log(ERROR, err.Error())
	}

	// 程序结束后关闭logger
	defer logger.Close()

	// 实现日志轮转
	// setupSignalHandler(logger)

	// 模拟日志写入
	for i := 0; i < 100; i++ {
		logger.Log(DEBUG, "This is a log message")
		time.Sleep(1 * time.Second)

		// if i == 20 {
		// 	// 向当前进程发送 SIGHUP
		// 	err := syscall.Kill(os.Getpid(), syscall.SIGHUP)
		// 	if err != nil {
		// 		log.Fatal("Failed to send SIGHUP:", err)
		// 	}
		// }

		logger.CheckRotate(cfg)
	}

	os.Exit(code)
}

// 实现日志轮转
func setupSignalHandler(logger *Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
	)

	go func() {
		for sig := range sigChan {
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				log.Println("Shutting down...")
				_ = logger.Close()
				os.Exit(0)

			case syscall.SIGHUP:
				filename := "log" + time.Now().Format("2006-01-02")
				log.Println("Received SIGHUP, rotating logs...")
				if err := logger.Reopen(filename); err != nil {
					log.Printf("Failed to rotate logs: %v", err)
				}
			}
		}
	}()
}
