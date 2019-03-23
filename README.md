# Golang bindings for the CoolQ HTTP API

[![GoDoc](https://godoc.org/github.com/catsworld/qq-bot-api?status.svg)](https://godoc.org/github.com/catsworld/qq-bot-api)
[![Build Status](https://travis-ci.org/catsworld/qq-bot-api.svg?branch=master)](https://travis-ci.org/catsworld/qq-bot-api)

If you're familiar with [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api),
you should quickly get familiar with this. -- Method names and architectures are just like the Telegram Bot API.

Head through the following examples and [godoc](https://godoc.org/github.com/catsworld/qq-bot-api) will give you a tutorial about how to use this package.
If you still have problems, look up to the code or open an issue.

## Communication Methods

CoolQ HTTP supports several communication methods.
The table below shows whether this SDK support a method.

+ √ means a method supported by both this SDK and CoolQ HTTP.  
+ × means a method supported by CoolQ HTTP but NOT by this SDK.  
+ \- means a method NOT supported by CoolQ HTTP.


| Methods | API | Event |
| --- | --- | --- |
| HTTP | √ | √ * |
| WebHook (HTTP Reverse) | - | √ |
| WebSocket | √ | √ |
| WebSocket Reverse | × | √ |

\* [CQHTTP LongPolling Plugin](https://github.com/richardchien/cqhttp-ext-long-polling) is required to support this feature.

## Quick Guide

This is a very simple bot that just displays any gotten updates, then replies it to that chat.

```go
func main() {
	bot, err := qqbotapi.NewBotAPI("MyCoolqHttpToken", "http://localhost:5700", "CQHTTP_SECRET")
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	u := qqbotapi.NewWebhook("/webhook_endpoint")
	u.PreloadUserInfo = true

	// use WebHook as event method
	updates := bot.ListenForWebhook(u)
	// or if you love WebSocket Reverse
	// updates := bot.ListenForWebSocket(u)
	go http.ListenAndServe("0.0.0.0:8443", nil)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.String(), update.Message.Text)

		bot.SendMessage(update.Message.Chat.ID, update.Message.Chat.Type, update.Message.Text)
	}
}
```

If you need to utilize a sync response, it will be slightly different.

```go
func main() {
	bot, err := qqbotapi.NewBotAPI("MyCoolqHttpToken", "http://localhost:5700", "CQHTTP_SECRET")
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	u := qqbotapi.NewWebhook("/webhook_endpoint")
	u.PreloadUserInfo = true
	bot.ListenForWebhookSync(u, func(update qqbotapi.Update) interface{} {

		log.Printf("[%s] %s", update.Message.From.String(), update.Message.Text)

		return map[string]interface{}{
			"reply": update.Message.Text,
		}
	})

	http.ListenAndServe("0.0.0.0:8443", nil)
}
```

It's as easy as well if you prefer WebSocket or LongPolling as event method.

```go
func main() {
	// whether to use WebSocket or LongPolling depends on the address
	// to use WebSocket, the address should be something like "ws://localhost:6700"
	bot, err := qqbotapi.NewBotAPI("MyCoolqHttpToken", "http://localhost:5700", "CQHTTP_SECRET")
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	u := qqbotapi.NewUpdate(0)
	u.PreloadUserInfo = true
	updates, err := bot.GetUpdatesChan(u)
	
	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.String(), update.Message.Text)

		bot.SendMessage(update.Message.Chat.ID, update.Message.Chat.Type, update.Message.Text)
	}
}
```


## Messages

Update.Message.Message is a group of Media, defined in package cqcode.

```go
	for update := range updates {
		if update.Message == nil {
			continue
		}

		for _, media := range *update.Message.Message {
			switch m := media.(type) {
			case *cqcode.Image:
				fmt.Printf("The message includes an image, id: %s, url: %s", m.FileID, m.URL)
			}
		}
	}
```

There are also some useful command helpers.

```go
	for update := range updates {
		if update.Message == nil {
			continue
		}

		// If this is true, a valid command must start with a command prefix (default to "/"), false by default.
		cqcode.StrictCommand = true
		// Set command prefix
		cqcode.CommandPrefix = "/"

		if update.Message.IsCommand() {
			// cmd string, args []string
			// In a StrictCommand mode, the command prefix will be stripped off.
			cmd, args := update.Message.Command()

			// Note that cmd and args is still media
			cmdMedia, _ := cqcode.ParseMessage(cmd)
			for _, v := range cmdMedia {
				switch v.(type) {
				case *cqcode.At:
					fmt.Print("The command includes an At!")
				case *cqcode.Face:
					fmt.Print("The command includes a Face!")
				}
			}
		}
	}
```

## Send Messages

The easiest way to send a message is to use a chained api.

```go
	// Send a text-img message
	m, err := bot.NewMessage(10000000, "group").
		At("1232332333").
		Text("嘤嘤嘤").
		NewLine().
		FaceByName("调皮").
		Text("这是一个测试").
		ImageBase64("img.jpg").
		Send()

	// Withdraw that message
	if err == nil {
		bot.DeleteMessage(m.MessageID)
	}

	// Send a stand-alone message (No need to call Send())
	bot.NewMessage(10000000, "private").
		Dice()
```

You can also use bot.SendMessage.

```go
	// All media types defined in package cqcode can be sent directly.
	// e.g. Send a text message
	bot.SendMessage(10000000, "group", cqcode.Text{
		Text: "[<- These will be encoded ->]",
	})

	// Send a location
	bot.SendMessage(10000000, "group", cqcode.Location{
		Content:   "上海市徐汇区交通大学华山路1954号",
		Latitude:  31.198878,
		Longitude: 121.436381,
		Style:     1,
		Title:     "位置分享",
	})

	// Send a message that contains a number of media.
	message := make(cqcode.Message, 0)
	message.Append(&cqcode.At{QQ: "all"})
	message.Append(&cqcode.Text{Text:" 大家起来嗨"})
	face, _ := cqcode.NewFaceFromName("调皮")
	message.Append(face)
	bot.SendMessage(10000000, "group", message)

	// To send an image or a record, you may use a helper function.
	// Format a base64-encoded image (Recommended)
	image1, err := qqbotapi.NewImageBase64("/path/to/image.jpg")

	// Format an image in the web.
	u, err := url.Parse("https://img.rikako.moe/i/D1D.jpg")
	image2 := qqbotapi.NewImageWeb(u)
	image2.DisableCache()

	// Format a local image if CQHTTP and your bot are under the same host
	// Have a look at https://github.com/catsworld/qq-bot-api/issues/11 if you do want to use this.
	u, err = url.Parse("file:///tmp/D1D.jpg")
	image3 := qqbotapi.NewImageWeb(u)
```

Or you can manually use the function bot.Send and bot.Do with a config.
You should find this quite familiar if you have once developed a Telegram bot.

```go
	// Alternaive to bot.SendMessage and bot.DeleteMessage
	message := qqbotapi.NewMessage(10000000, "group", "aaaaaa")
	m, err := bot.Send(message)
	if err == nil {
		config := qqbotapi.DeleteMessageConfig{MessageID: m.MessageID}
		bot.Do(config)
	}
```
