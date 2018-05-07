package qqbotapi

import (
	"encoding/json"
	"net/url"
	"strconv"
)

// Library errors
const (
	ErrBadURL = "bad or empty url"
)

// Chattable is any config type that can be sent.
type Chattable interface {
	values() (url.Values, error)
	method() string
}

// BaseChat is base type for all chat config types.
type BaseChat struct {
	ChatID              int64 // required
	ChannelUsername     string
	ReplyToMessageID    int
	ReplyMarkup         interface{}
	DisableNotification bool
}

// values returns url.Values representation of BaseChat
func (chat *BaseChat) values() (url.Values, error) {
	v := url.Values{}
	if chat.ChannelUsername != "" {
		v.Add("chat_id", chat.ChannelUsername)
	} else {
		v.Add("chat_id", strconv.FormatInt(chat.ChatID, 10))
	}

	if chat.ReplyToMessageID != 0 {
		v.Add("reply_to_message_id", strconv.Itoa(chat.ReplyToMessageID))
	}

	if chat.ReplyMarkup != nil {
		data, err := json.Marshal(chat.ReplyMarkup)
		if err != nil {
			return v, err
		}

		v.Add("reply_markup", string(data))
	}

	v.Add("disable_notification", strconv.FormatBool(chat.DisableNotification))

	return v, nil
}

// MessageConfig contains information about a SendMessage request.
type MessageConfig struct {
	BaseChat
	SendType              string
	Text                  string
	ParseMode             string
	DisableWebPagePreview bool
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
	Offset  int
	Limit   int
	Timeout int
}

// ChatConfig contains information about getting information on a chat.
type ChatConfig struct {
	ChatID             int64
	SuperGroupUsername string
}

// ChatConfigWithUser contains information about getting information on
// a specific user within a chat.
type ChatConfigWithUser struct {
	ChatID             int64
	SuperGroupUsername string
	UserID             int
}
