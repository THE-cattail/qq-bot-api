package qqbotapi

import (
	"net/url"
	"strconv"
)

// Chattable is any config type that can be sent.
type Chattable interface {
	values() (url.Values, error)
	method() string
}

// BaseChat is base type for all chat config types.
type BaseChat struct {
	ChatID int64 // required
}

// values returns url.Values representation of BaseChat
func (chat *BaseChat) values() (url.Values, error) {
	v := url.Values{}
	v.Add("chat_id", strconv.FormatInt(chat.ChatID, 10))

	return v, nil
}

// MessageConfig contains information about a SendMessage request.
type MessageConfig struct {
	BaseChat
	SendType string
	Text     string
}

// values returns a url.Values representation of MessageConfig.
func (config MessageConfig) values() (url.Values, error) {
	v, err := config.BaseChat.values()
	if err != nil {
		return v, err
	}
	v.Add("message_type", config.SendType)
	v.Add("user_id", strconv.FormatInt(config.BaseChat.ChatID, 10))
	v.Add("group_id", strconv.FormatInt(config.BaseChat.ChatID, 10))
	v.Add("discuss_id", strconv.FormatInt(config.BaseChat.ChatID, 10))
	v.Add("message", config.Text)

	return v, nil
}

// method returns Telegram API method name for sending Message.
func (config MessageConfig) method() string {
	return "send_msg"
}

// UpdateConfig contains information about a GetUpdates request.
type UpdateConfig struct {
	BaseUpdateConfig
	Offset  int
	Limit   int
	Timeout int
}

type WebhookConfig struct {
	BaseUpdateConfig
	Pattern string
}

type BaseUpdateConfig struct {
	PreloadUserInfo bool
}
