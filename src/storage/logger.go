package storage

import (
	"PromotingEfficiency/src/config"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LogLevel 定义日志级别类型
type LogLevel int

// 日志级别常量定义
const (
	DEBUG   LogLevel = iota // 调试信息
	INFO                    // 普通信息
	WARNING                 // 警告信息
	ERROR                   // 错误信息
	FATAL                   // 致命错误
)

// Logger 日志记录器结构体
type Logger struct {
	file        *os.File      // 日志文件句柄
	mu          sync.Mutex    // 互斥锁，保证并发安全
	subscribers []chan string // 订阅者通道列表
}

// NewLogger 创建新的日志记录器
// 参数:
//
//	filename: 日志文件路径
//
// 返回值:
//
//	*Logger: 日志记录器实例
//	error: 创建过程中的错误
func NewLogger(filename string) (*Logger, error) {
	// 打开或创建日志文件，权限设置为0644
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		file: file,
	}, nil
}

// Log 关闭
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Reopen 重新打开一个文件
// 参数：
// filename：新文件的路径
// 返回值：
// error：重建文件时的错误
func (l *Logger) Reopen(filename string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 关闭旧文件
	if l.file != nil {
		_ = l.file.Close()
	}

	// 重新打开
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	l.file = file
	return nil
}

// Log 记录日志方法
// 参数:
//
//	level: 日志级别
//	message: 日志消息内容
func (l *Logger) Log(level LogLevel, message string) {
	l.mu.Lock()         // 加锁保证线程安全
	defer l.mu.Unlock() // 方法结束时自动解锁

	// 格式化日志条目: [时间] 级别: 消息
	entry := fmt.Sprintf("[%s] %s: %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		level.String(),
		message)

	// 写入日志文件
	l.file.WriteString(entry)

	// 通知所有订阅者
	for _, ch := range l.subscribers {
		select {
		case ch <- entry: // 尝试发送日志条目
		default: // 如果通道已满则跳过
		}
	}
}

func (l *Logger) CheckRotate(cfg *config.Config) {
	info, err := l.file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	if info.Size() > eval(cfg.LogMaxSize) {
		l.rotateLog()
	}
}

func (l *Logger) rotateLog() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.file.Close()
		os.Rename("app.log", fmt.Sprintf("app.%s.log", time.Now().Format("20060102150405")))
	}

	var err error
	l.file, err = os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(l.file)
}

// Subscribe 订阅日志消息
// 返回值:
//
//	<-chan string: 只读通道，用于接收日志消息
func (l *Logger) Subscribe() <-chan string {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 创建带缓冲的通道(容量100)
	ch := make(chan string, 100)
	// 将新通道加入订阅者列表
	l.subscribers = append(l.subscribers, ch)
	return ch
}

// String 实现LogLevel的String方法
// 返回值:
//
//	string: 日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

func eval(expr string) int64 {
	parts := strings.Split(expr, " * ")
	var result int64 = 1
	for _, part := range parts {
		num, _ := strconv.Atoi(part)
		result *= int64(num)
	}
	return result
}

// 以下是快捷日志方法
func (l *Logger) Debug(msg string)   { l.Log(DEBUG, msg) }   // 记录调试信息
func (l *Logger) Info(msg string)    { l.Log(INFO, msg) }    // 记录普通信息
func (l *Logger) Warning(msg string) { l.Log(WARNING, msg) } // 记录警告信息
func (l *Logger) Error(msg string)   { l.Log(ERROR, msg) }   // 记录错误信息
func (l *Logger) Fatal(msg string)   { l.Log(FATAL, msg) }   // 记录致命错误
