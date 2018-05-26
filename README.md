# Golang bindings for the Coolq HTTP API

[![GoDoc](https://godoc.org/github.com/catsworld/qq-bot-api?status.svg)](https://godoc.org/github.com/catsworld/qq-bot-api)
[![Build Status](https://travis-ci.org/catsworld/qq-bot-api.svg?branch=master)](https://travis-ci.org/catsworld/qq-bot-api)

## 示例

这是一个非常简单的复读机器人。

```go
func main() {
	bot, err := qqbotapi.NewBotAPI("MyCoolqHttpToken", "http://localhost:5700", "")
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	u := qqbotapi.NewWebhook("/webhook_endpoint")
	u.PreloadUserInfo = true
	updates := bot.ListenForWebhook(u)
	go http.ListenAndServe("0.0.0.0:8443", nil)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.String(), update.Message.Text)

		msg := qqbotapi.NewMessage(update.Message.Chat.ID, update.Message.Chat.Type, update.Message.Text)
		bot.Send(msg)
	}
}
```