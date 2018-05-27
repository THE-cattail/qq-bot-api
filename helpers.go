package qqbotapi

import (
	"fmt"
	"github.com/catsworld/qq-bot-api/cqcode"
)

// NewMessage creates a new Message.
//
// chatID is where to send it, message is the message.
func NewMessage(chatID int64, sendType string, message interface{}) MessageConfig {
	mc := MessageConfig{
		BaseChat: BaseChat{
			ChatID:   chatID,
			SendType: sendType,
		},
	}
	switch v := message.(type) {
	case cqcode.Message:
		mc.Text = v.CQString()
	case cqcode.Media:
		mc.Text = cqcode.FormatCQCode(v)
	case string:
		mc.Text = v
	default:
		mc.Text = fmt.Sprint(v)
	}
	return mc
}

// NewUpdate gets updates since the last Offset.
//
// offset is the last Update ID to include.
// You likely want to set this to the last Update ID plus 1.
func NewUpdate(offset int) UpdateConfig {
	return UpdateConfig{
		BaseUpdateConfig: BaseUpdateConfig{
			PreloadUserInfo: false,
		},
		Offset:  offset,
		Limit:   0,
		Timeout: 0,
	}
}

func NewWebhook(pattern string) WebhookConfig {
	return WebhookConfig{
		BaseUpdateConfig: BaseUpdateConfig{
			PreloadUserInfo: false,
		},
		Pattern: pattern,
	}
}
