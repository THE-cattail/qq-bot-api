package qqbotapi

import (
	"encoding/json"
)

// APIResponse is a response from the Coolq HTTP API with the result
// stored raw.
type APIResponse struct {
	Status  string          `json:"status"`
	Data    json.RawMessage `json:"data"`
	RetCode int64           `json:"retcode"`
}

// Update is an update response, from GetUpdates.
type Update struct {
	UpdateID    int      `json:"update_id"`
	PostType    string   `json:"post_type"`
	MessageType string   `json:"message_type"`
	MessageID   int      `json:"message_id"`
	GroupID     int      `json:"group_id"`
	DiscussID   int      `json:"discuss_id"`
	UserID      int      `json:"user_id"`
	Text        string   `json:"message"`
	Message     *Message `json:"-"`
}

// UpdatesChannel is the channel for getting updates.
type UpdatesChannel <-chan Update

// User is a user on QQ.
type User struct {
	ID       int    `json:"user_id"`
	NickName string `json:"nickname"`
	Card     string `json:"card"`
	Title    string `json:"title"`
}

// String displays a simple text version of a user.
//
// It is normally a user's card, but falls back to a nickname as available.
func (u *User) String() string {
	p := ""
	if u.Title != "" {
		p = "[" + u.Title + "]"
	}
	if u.Card != "" {
		return p + u.Card
	}
	return p + u.NickName
}

// Chat contains information about the place a message was sent.
type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

// IsPrivate returns if the Chat is a private conversation.
func (c Chat) IsPrivate() bool {
	return c.Type == "private"
}

// IsGroup returns if the Chat is a group.
func (c Chat) IsGroup() bool {
	return c.Type == "group"
}

// IsDiscuss returns if the Chat is a discuss.
func (c Chat) IsDiscuss() bool {
	return c.Type == "discuss"
}

// Message is returned by almost every request, and contains data about
// almost anything.
type Message struct {
	MessageID int    `json:"message_id"`
	From      *User  `json:"from"` // optional
	Chat      *Chat  `json:"chat"`
	Text      string `json:"text"` // optional
}
