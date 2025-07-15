// client.go
package email

import (
	// 标准库导入
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"mime"
	"net/smtp"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	// 第三方库导入
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/jordan-wright/email"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	// 项目内部导入
	"PromotingEfficiency/src/config"
	"PromotingEfficiency/src/storage"
)

/******************** 常量定义 ********************/
const (
	MaxFetchMessages   = 100            // 单次最大获取邮件数量，防止内存溢出
	FetchBufferSize    = 10             // 邮件获取通道缓冲区大小
	RecentMailDuration = 24 * time.Hour // 判定为"新邮件"的时间范围
)

/******************** 接口定义 ********************/

// MailService 邮件服务核心接口
type MailService interface {
	// Connect 建立与邮件服务器的连接
	// 返回: 连接错误信息
	Connect() error

	// Disconnect 安全断开与邮件服务器的连接
	Disconnect()

	// FetchUnreadEmails 获取未读邮件列表
	// 返回: 邮件列表，错误信息
	FetchUnreadEmails() ([]*Email, error)
}

// EmailHandler 邮件处理器接口
type EmailHandler interface {
	// Handle 处理单个邮件
	// 参数: email - 待处理的邮件对象
	// 返回: 处理过程中的错误
	Handle(email *Email) error
}

/******************** 数据结构 ********************/

// Email 邮件基础数据结构
type Email struct {
	UID         uint32        // 邮件唯一标识符(IMAP UID)
	Date        time.Time     // 邮件发送时间
	From        string        // 发件人信息(已解码)
	Subject     string        // 邮件主题(已解码)
	Attachments []*Attachment // 邮件附件列表
}

// Attachment 邮件附件数据结构
type Attachment struct {
	Filename string // 附件文件名(已解码)
	Content  []byte // 附件二进制内容
}

/******************** 邮件客户端实现 ********************/

// EmailClient IMAP邮件客户端实现
type EmailClient struct {
	server    string         // IMAP服务器地址(包含端口)
	username  string         // 登录用户名
	password  string         // 登录密码/授权码
	client    *client.Client // IMAP客户端实例
	mu        sync.Mutex     // 线程安全锁
	connected bool           // 连接状态标记
}

// NewEmailClient 构造函数：创建邮件客户端实例
// 参数:
//   - server: 服务器地址(如"imap.qq.com:993")
//   - username: 邮箱账号
//   - password: 密码/授权码
//
// 返回: 初始化后的邮件客户端指针
func NewEmailClient(server, username, password string) *EmailClient {
	return &EmailClient{
		server:   server,
		username: username,
		password: password,
	}
}

// Connect 建立安全连接(线程安全)
// 实现流程:
// 1. 检查现有连接有效性
// 2. 创建带超时的TLS连接
// 3. 执行登录认证
// 4. 更新连接状态
func (s *EmailClient) Connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 连接有效性检查
	if s.connected {
		if _, err := s.client.Capability(); err == nil {
			return nil
		}
		// 连接已失效则重置
		s.client.Logout()
		s.client = nil
	}

	c, err := client.DialTLS(s.server, nil)
	if err != nil {
		return fmt.Errorf("连接服务器失败: %w", err)
	}

	// 登录认证
	if err := c.Login(s.username, s.password); err != nil {
		c.Logout()
		return fmt.Errorf("登录失败: %w", err)
	}

	// 更新状态
	s.client = c
	s.connected = true
	return nil
}

// Disconnect 安全断开连接(线程安全)
func (s *EmailClient) Disconnect() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil {
		s.client.Logout()
		s.client = nil
	}
	s.connected = false
}

func SendEmail(c *config.Config) {
	from := c.SendEmail.Username
	password := c.SendEmail.Password
	to := []string{c.Email.Username}

	e := email.NewEmail()
	e.From = fmt.Sprintf("Sender <%s>", from)
	e.To = to
	e.Subject = c.Email.TargetSubject
	e.Text = []byte("这是一封测试邮件，内容与兴效能相关。")

	// 添加附件
	if attachmentPath := c.SendEmail.Attachment; attachmentPath != "" {
		if _, err := os.Stat(attachmentPath); err == nil {
			if _, err := e.AttachFile(attachmentPath); err != nil {
				log.Printf("附件添加失败: %v", err)
			}
		} else {
			log.Printf("附件文件不存在: %s", attachmentPath)
		}
	}

	// 确保服务器地址包含端口
	smtpAddr := c.SendEmail.Server
	if !strings.Contains(smtpAddr, ":") {
		smtpAddr += ":465" // 默认 SSL 端口
	}

	// 发送邮件（显式 TLS）
	err := e.SendWithTLS(
		smtpAddr,
		smtp.PlainAuth("", from, password, strings.Split(smtpAddr, ":")[0]),
		&tls.Config{ServerName: strings.Split(smtpAddr, ":")[0]},
	)
	if err != nil {
		log.Fatalf("邮件发送失败: %v (Server: %s)", err, smtpAddr)
	}
	log.Println("邮件发送成功！")
}

// FetchUnreadEmails 获取未读邮件(线程安全)
// 实现逻辑:
// 1. 检查连接状态
// 2. 选择INBOX邮箱
// 3. 设置搜索条件(未读+24小时内)
// 4. 执行搜索并获取邮件内容
func (s *EmailClient) FetchUnreadEmails() ([]*Email, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.connected {
		return nil, fmt.Errorf("未连接到邮件服务器")
	}

	// 选择收件箱
	if _, err := s.client.Select("INBOX", false); err != nil {
		return nil, fmt.Errorf("选择邮箱失败: %w", err)
	}

	// 构建搜索条件
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	criteria.Since = time.Now().Add(-RecentMailDuration)

	// 执行搜索
	ids, err := s.client.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("搜索邮件失败: %w", err)
	}

	// 空结果处理
	if len(ids) == 0 {
		return nil, nil
	}

	// 限制获取数量
	if len(ids) > MaxFetchMessages {
		ids = ids[:MaxFetchMessages]
	}

	return s.fetchMessages(ids)
}

// fetchMessages 获取指定ID的邮件内容
// 参数:
//   - ids: 邮件ID序列
//
// 返回: 解析后的邮件列表
func (s *EmailClient) fetchMessages(ids []uint32) ([]*Email, error) {
	// 准备获取请求
	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{
		imap.FetchEnvelope,     // 信封信息(发件人、主题等)
		imap.FetchFlags,        // 邮件标志
		imap.FetchInternalDate, // 内部日期
		imap.FetchUid,          // 唯一标识
		section.FetchItem(),    // 正文内容
	}

	// 异步获取通道
	messages := make(chan *imap.Message, FetchBufferSize)
	done := make(chan error, 1)

	// 启动异步获取
	go func() {
		done <- s.client.Fetch(seqset, items, messages)
	}()

	// 处理获取结果
	var emails []*Email
	for msg := range messages {
		email, err := s.parseEmail(msg, section)
		if err != nil {
			log.Printf("解析邮件失败: %v", err)
			continue
		}
		emails = append(emails, email)
	}

	// 检查获取错误
	if err := <-done; err != nil {
		return nil, fmt.Errorf("获取邮件内容失败: %w", err)
	}

	return emails, nil
}

/******************** 邮件解析相关 ********************/

// parseEmail 解析单个邮件
// 参数:
//   - msg: 原始邮件数据
//   - section: 正文部分标识
//
// 返回: 解析后的邮件对象
func (s *EmailClient) parseEmail(msg *imap.Message, section *imap.BodySectionName) (*Email, error) {
	r := msg.GetBody(section)
	if r == nil {
		return nil, fmt.Errorf("邮件正文为空")
	}

	// 创建邮件阅读器
	mr, err := mail.CreateReader(r)
	if err != nil {
		return nil, fmt.Errorf("创建邮件阅读器失败: %w", err)
	}

	// 解析基础信息
	header := mr.Header
	date, _ := header.Date() // 日期解析错误不影响后续处理

	email := &Email{
		UID:     msg.Uid,
		Date:    date,
		From:    decodeHeader(header.Get("From")),
		Subject: decodeHeader(header.Get("Subject")),
	}

	// 解析邮件各部分
	if err := s.parseEmailParts(mr, email); err != nil {
		return nil, err
	}

	return email, nil
}

// parseEmailParts 解析邮件正文和附件
// 参数:
//   - mr: 邮件阅读器
//   - email: 待补充的邮件对象
func (s *EmailClient) parseEmailParts(mr *mail.Reader, email *Email) error {
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // 跳过解析失败的部分
		}

		// 处理附件部分
		if h, ok := p.Header.(*mail.AttachmentHeader); ok {
			if err := s.parseAttachment(h, p.Body, email); err != nil {
				log.Printf("解析附件失败: %v", err)
			}
		}
	}
	return nil
}

// parseAttachment 解析单个附件
// 参数:
//   - h: 附件头信息
//   - body: 附件内容流
//   - email: 所属邮件对象
func (s *EmailClient) parseAttachment(h *mail.AttachmentHeader, body io.Reader, email *Email) error {
	filename, err := h.Filename()
	if err != nil || filename == "" {
		return fmt.Errorf("无效的附件名")
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, body); err != nil {
		return fmt.Errorf("读取附件内容失败: %w", err)
	}

	email.Attachments = append(email.Attachments, &Attachment{
		Filename: decodeHeader(filename),
		Content:  buf.Bytes(),
	})
	return nil
}

/******************** 工具函数 ********************/

// decodeHeader 解码邮件头特殊编码
// 支持格式: =?charset?encoding?encoded-text?=
func decodeHeader(header string) string {
	decoder := mime.WordDecoder{
		CharsetReader: charsetReader,
	}

	decoded, err := decoder.DecodeHeader(header)
	if err != nil {
		return header // 解码失败返回原始内容
	}
	return decoded
}

// charsetReader 字符集转换器
// 支持GBK/GB2312自动转UTF-8
func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	charset = strings.ToLower(charset)
	switch charset {
	case "gbk", "gb2312":
		return transform.NewReader(input, simplifiedchinese.GBK.NewDecoder()), nil
	default:
		return input, nil // 其他编码原样返回
	}
}

/******************** 业务逻辑函数 ********************/

// CheckAndProcessEmails 邮件处理主流程
// 参数:
//   - mailService: 邮件服务实例
//   - handler: 邮件处理器
//   - logger: 日志记录器
//
// 返回: 处理过程中的错误
func CheckAndProcessEmails(mailService MailService, logger *storage.Logger) (email *Email, err error) {
	startTime := time.Now()
	logger.Info("开始检查邮箱...")

	// 建立连接
	if err := mailService.Connect(); err != nil {
		return email, fmt.Errorf("连接失败: %w", err)
	}
	defer mailService.Disconnect() // 确保连接关闭

	// 获取未读邮件
	emails, err := mailService.FetchUnreadEmails()
	if err != nil {
		return email, fmt.Errorf("获取邮件失败: %w", err)
	}

	// 空结果处理
	if len(emails) == 0 {
		logger.Info("没有新邮件")
		return email, err
	}

	// 过滤目标邮件
	targetEmail := filterLatestTargetEmail(emails, "兴效能")
	if targetEmail == nil {
		logger.Info("没有目标邮件")
		return email, err
	}

	logger.Info(fmt.Sprintf("处理完成，耗时: %v", time.Since(startTime)))
	return targetEmail, nil
}

// filterLatestTargetEmail 过滤符合条件的最近邮件
// 参数:
//   - emails: 待过滤邮件列表
//   - keyword: 主题关键词
//
// 返回: 最新目标邮件(按日期排序)
func filterLatestTargetEmail(emails []*Email, keyword string) *Email {
	var targetEmails []*Email
	for _, email := range emails {
		if strings.Contains(email.Subject, keyword) {
			targetEmails = append(targetEmails, email)
		}
	}

	if len(targetEmails) == 0 {
		return nil
	}

	// 按日期降序排序
	sort.Slice(targetEmails, func(i, j int) bool {
		return targetEmails[i].Date.After(targetEmails[j].Date)
	})

	return targetEmails[0]
}
