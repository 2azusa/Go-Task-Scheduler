package notify

import (
	"bytes"
	"fmt"
	"html/template"
	"pulse/common/pkg/logger"

	"github.com/go-gomail/gomail"
)

var _defaultMail *Mail

type Mail struct {
	Port     int
	From     string
	Host     string
	Secret   string
	Nickname string
}

// SendMsg 方法用于发送一封邮件
func (mail *Mail) SendMsg(msg *Message) {
	m := gomail.NewMessage()

	m.SetHeader("From", m.FormatAddress(_defaultMail.From, _defaultMail.Nickname))
	m.SetHeader("To", msg.To...)
	m.SetHeader("Subject", msg.Subject)
	msgData := parseMailTemplate(msg) // 将模板和数据数据，生成HTML邮件正文
	m.SetBody("text/html", msgData)

	d := gomail.NewDialer(_defaultMail.Host, _defaultMail.Port, _defaultMail.From, _defaultMail.Secret)
	// 连接到SMTP服务器并发送邮件
	if err := d.DialAndSend(m); err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("smtp send msg[%+v] err: %s", msg, err.Error()))
	}
}

// 解析邮件模板并填充邮件
func parseMailTemplate(msg *Message) string {
	tmpl, err := template.New("notify").Parse(mailTemplate)
	if err != nil {
		return fmt.Sprintf("Failed to parse the notification template error: %s", err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, msg)
	if err != nil {
		return fmt.Sprintf("Failed to parse the notification template execute error: %s", err.Error())
	}
	return buf.String()
}

// 存储了邮件的HTML模板
var mailTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
    <title></title>
    <meta charset="utf-8"/>

</head>
<body>
<div class="cap" style="
            border: 2px solid black;
            background-color: whitesmoke;
            height: 500px"
>
    <div class="content" style="
            background-color: white;
            background-clip: padding-box;
            color:black;
            font-size: 13px;
            padding: 25px 25px;
            margin: 25px 25px;
        ">
        <div class="hello" style="text-align: center; color: #FF3333;font-size: 18px;font-weight: bolder">
            {{.Subject}}
        </div>
        <br>
        <div>
            <table border="1"  bordercolor="black" cellspacing="0px" cellpadding="4px" style="margin: 0 auto;">
                <tr >
                    <td>告警主机</td>
                    <td >{{.IP}}</td>
                </tr>

                <tr>
                    <td>告警时间</td>
                    <td>{{.OccurTime}}</td>
                </tr>

                <tr>
                    <td>告警信息</td>
                    <td>{{.Body}}</td>
                </tr>

            </table>
        </div>
        <br><br>
    </div>
</div>
<br>

</body>
</html>
`
