package notify

import (
	"pulse/common/pkg/utils"
	"strings"
	"time"
)

// 定义通知类型的常量
const (
	NotifyTypeMail    = 1
	NotifyTypeWebHook = 2
)

type Message struct {
	Type      int
	To        []string
	Subject   string
	Body      string
	IP        string // 相关的IP地址
	OccurTime string // 发送时间
}

type Noticer interface {
	SendMsg(*Message)
}

var msgQueue chan *Message // 消息队列

func Init(mail *Mail, web *WebHook) {
	_defaultMail = &Mail{
		Port:     mail.Port,
		From:     mail.From,
		Host:     mail.Host,
		Secret:   mail.Secret,
		Nickname: mail.Nickname,
	}
	_defaultWebHook = &WebHook{
		Kind: web.Kind,
		Url:  web.Url,
	}

	msgQueue = make(chan *Message, 64)
}

// 发送一条通知
func Send(msg *Message) {
	msgQueue <- msg
}

// 从消息队列读取消息，并分发给相应的处理器
func Serve() {
	for msg := range msgQueue {
		if msg == nil {
			continue
		}

		switch msg.Type {
		case NotifyTypeMail:
			msg.Check()
			_defaultMail.SendMsg(msg)
		case NotifyTypeWebHook:
			msg.Check()
			go _defaultWebHook.SendMsg(msg) // 异步调用
		}
	}
}

func (m *Message) Check() {
	if m.OccurTime == "" {
		m.OccurTime = time.Now().Format(utils.TimeFormatSecond)
	}

	m.Body = strings.ReplaceAll(m.Body, "\"", "'")
	m.Body = strings.ReplaceAll(m.Body, "\n", "")
}
