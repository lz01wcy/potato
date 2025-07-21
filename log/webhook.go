package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
	"time"
)

type Notification struct {
	MsgType string               `json:"msg_type"`
	Content *NotificationContent `json:"content"`
}

type NotificationContent struct {
	Text string `json:"text"`
}

// SetErrorNotification 设置错误通知
// url: webhook url
// keyword: 通知关键字 用于机器人过滤信息
func SetErrorNotification(url, keyword string) {
	errCh := make(chan string, 64)
	Logger = Logger.WithOptions(zap.Hooks(func(entry zapcore.Entry) error {
		if entry.Level >= zap.ErrorLevel {
			exec, _ := os.Executable()
			msg := fmt.Sprintf("%s%s %s\n%s\n%s\n%s", "[Error]", keyword, entry.Time.Format("2006-01-02 15:04:05.000"), exec, entry.Message, entry.Stack)
			errCh <- msg
		}
		return nil
	}))
	Sugar = Logger.Sugar()
	go runPushNotification(url, errCh)
}

// 推送进程
func runPushNotification(url string, ch chan string) {
	defer func() {
		if r := recover(); r != nil {
			Sugar.Errorf("%s\n\n", r)
		}
	}()
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	for msg := range ch {
		ntf := &Notification{
			MsgType: "text",
			Content: &NotificationContent{Text: msg},
		}
		data, _ := json.Marshal(ntf)
		contentType := "application/json"
		resp, err := client.Post(url, contentType, bytes.NewReader(data))
		if err != nil {
			fmt.Printf("runPushNotification: %s error: %v\n", msg, err)
		} else {
			_ = resp.Body.Close()
		}
	}
}
