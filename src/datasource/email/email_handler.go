// email_handler.go
package email

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ====================== 邮件处理器实现 ======================

type XLSXAttachmentHandler struct {
	TargetSubject string          // 目标邮件主题关键词
	DataDir       string          // 附件保存目录
	processedUIDs map[uint32]bool // 已处理邮件UID记录
	mu            sync.RWMutex    // 保护processedUIDs的读写锁
}

func NewXLSXAttachmentHandler(subject, dataDir string) *XLSXAttachmentHandler {
	// 初始化processedUIDs映射

	return &XLSXAttachmentHandler{
		TargetSubject: subject,
		DataDir:       dataDir,
		processedUIDs: make(map[uint32]bool), // 初始化映射
	}
}

// isProcessed 检查邮件是否已处理过（线程安全）
func (h *XLSXAttachmentHandler) isProcessed(uid uint32) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.processedUIDs[uid]
}

// markAsProcessed 标记邮件为已处理（线程安全）
func (h *XLSXAttachmentHandler) markAsProcessed(uid uint32) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.processedUIDs[uid] = true
}

// Handle 处理单个邮件
func (h *XLSXAttachmentHandler) Handle(email *Email) error {
	// 检查是否已处理过该邮件
	if h.isProcessed(email.UID) {
		return nil
	}

	// 检查邮件主题是否包含目标关键词
	if !strings.Contains(email.Subject, h.TargetSubject) {
		fmt.Printf("跳过主题不匹配的邮件: %s\n", email.Subject)
		return nil
	}

	// 打印邮件基本信息
	fmt.Printf("\n处理邮件: %s\n发件人: %s\n日期: %s\n",
		email.Subject, email.From, email.Date.Format("2006-01-02 15:04:05"))

	// 确保保存目录存在
	if err := os.MkdirAll(h.DataDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 处理附件
	hasXLSX := false
	for _, attachment := range email.Attachments {
		// 只处理XLSX文件
		if filepath.Ext(attachment.Filename) == ".xlsx" {
			fmt.Println("找到XLSX附件:", attachment.Filename)

			// 构建完整文件路径
			filePath := filepath.Join(h.DataDir, attachment.Filename)

			// 保存文件
			if err := os.WriteFile(filePath, attachment.Content, 0644); err != nil {
				return fmt.Errorf("保存附件失败: %v", err)
			}

			fmt.Printf("附件已保存到: %s\n", filePath)
			hasXLSX = true
		}
	}

	// 如果有XLSX附件，则标记邮件为已处理
	if hasXLSX {
		h.markAsProcessed(email.UID)
	}

	return nil
}
