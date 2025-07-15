package main

import (
	"PromotingEfficiency/src/config"
	"PromotingEfficiency/src/datasource/email"
	"PromotingEfficiency/src/storage"
	"fmt"
	"log"
	"testing"
)

func TestStartWebUI(t *testing.T) {

	// 初始化配置
	jsonFolder := "config"
	jsonFile := "config.json"
	cfg := config.LoadConfig(jsonFolder, jsonFile)
	fmt.Println(cfg)

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
	fmt.Println(mailClient)
}

// func TestMain1(m *testing.M) {
// 	// 测试前初始化
// 	setup()
// 	code := m.Run() // 运行测试

// 	// 测试后清理
// 	teardown()
// 	os.Exit(code)
// }

// func setup() {
// 	fmt.Println("Before all tests")
// }

// func teardown() {
// 	fmt.Println("After all tests")
// }
