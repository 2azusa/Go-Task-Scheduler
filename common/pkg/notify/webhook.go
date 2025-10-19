package notify

import (
	"encoding/json"
	"fmt"
	"pulse/common/pkg/httpclient"
	"pulse/common/pkg/logger"
	"strings"
)

var _defaultWebHook *WebHook

type WebHook struct {
	Kind string // 类型
	Url  string // 接收地址
}

func (w *WebHook) SendMsg(msg *Message) {
	switch _defaultWebHook.Kind {
	case "feishu":
		var sendData = feiShuTemplateCard
		sendData = strings.Replace(sendData, "timeSlot", msg.OccurTime, 1)
		sendData = strings.Replace(sendData, "ipSlot", msg.IP, 1)

		// 飞书卡片消息中，@某人需要 <at> 标签
		userSlot := ""
		for _, to := range msg.To {
			// 将每个收件人构成一个 <at> 标签拼接到字符串中
			userSlot += fmt.Sprintf("<at email='' >%s</at>", to)
		}
		sendData = strings.Replace(sendData, "userSlot", userSlot, 1)
		sendData = strings.Replace(sendData, "subjectSlot", msg.Subject, 1)
		sendData = strings.Replace(sendData, "msgSlot", msg.Body, 1)

		_, err := httpclient.PostJson(_defaultWebHook.Url, sendData, 0)
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("feishu send msg[%+v] err: %s", msg, err.Error()))
		}
	default:
		data, err := json.Marshal(msg)
		if err != nil {
			return
		}

		_, err = httpclient.PostJson(_defaultWebHook.Url, string(data), 0)
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("webhook api send msg[%+v] err: %s", msg, err.Error()))
		}
	}
}
