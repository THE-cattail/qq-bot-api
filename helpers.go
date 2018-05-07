package qqbotapi

// NewMessage creates a new Message.
//
// chatID is where to send it, text is the message text.
func NewMessage(chatID int64, sendType string, text string) MessageConfig {
	return MessageConfig{
		BaseChat: BaseChat{
			ChatID:           chatID,
			ReplyToMessageID: 0,
		},
		SendType: sendType,
		Text:     text,
		DisableWebPagePreview: false,
	}
}

// NewUpdate gets updates since the last Offset.
//
// offset is the last Update ID to include.
// You likely want to set this to the last Update ID plus 1.
func NewUpdate(offset int) UpdateConfig {
	return UpdateConfig{
		Offset:  offset,
		Limit:   0,
		Timeout: 0,
	}
}
