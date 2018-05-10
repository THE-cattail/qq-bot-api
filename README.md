# Golang bindings for the Coolq HTTP API

[![GoDoc](https://godoc.org/github.com/catsworld/qq-bot-api?status.svg)](https://godoc.org/github.com/catsworld/qq-bot-api)

基于 `github.com/go-telegram-bot-api/telegram-bot-api` 修改。

必须使用 `github.com/catsworld/cqhttp-longpoll-server` 搭建长轮询服务
（也可以自行搭建与 `github.com/jcuga/golongpoll` 发布数据格式相同的长轮询服务）

## 示例

这是一个非常简单的复读机器人。

```go
package main

import (
    "log"
    "github.com/catsworld/qq-bot-api"
)

func main() {
    bot, err := qqbotapi.NewBotAPI("MyAwesomeBotToken", "MyAPIEndpoint", "MyLongpollServerEndpoint")
    if err != nil {
        log.Panic(err)
    }

    bot.Debug = true

    log.Printf("Authorized on account %v", bot.Self.String())

    u := qqbotapi.NewUpdate(0)
    u.Timeout = 60

    updates, err := bot.GetUpdatesChan(u)

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