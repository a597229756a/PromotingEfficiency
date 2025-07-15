package file

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestMain(m *testing.M) {

}

func TestXxx(t *testing.T) {

	// 配置
	cfg := Config{
		FolderName:    "data",
		Level:         4,
		Keyword:       "兴效能",
		CheckInterval: 5 * time.Second,
	}

	// 创建文件监控器
	monitor, err := NewFileMonitor(cfg.FolderName)
	if err != nil {
		fmt.Printf("Failed to create file monitor: %v\n", err)
		return
	}

	fmt.Printf("Monitoring directory: %s for Excel files containing '%s'\n",
		cfg.FolderName, cfg.Keyword)

	// 设置上下文和信号处理
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	SetupSignalHandler(cancel)

	// 运行监控
	if err := monitor.Run(ctx); err != nil {
		fmt.Printf("Monitoring error: %v\n", err)
	}

	fmt.Println("File monitoring stopped")

}
