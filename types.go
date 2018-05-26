package qqbotapi

import (
	"encoding/json"
	"github.com/catsworld/qq-bot-api/cqcode"
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
	PostType      string      `json:"post_type"`
	MessageType   string      `json:"message_type"`
	SubType       string      `json:"sub_type"`
	MessageID     int         `json:"message_id"`
	GroupID       int         `json:"group_id"`
	DiscussID     int         `json:"discuss_id"`
	UserID        int         `json:"user_id"`
	Font          int         `json:"font"`
	RawMessage    interface{} `json:"message"`
	AnonymousName string      `json:"anonymous"`
	AnonymousFlag string      `json:"anonymous_flag"` // Anonymous ID
	Event         string      `json:"event"`
	OperatorID    int         `json:"operator_id"`
	File          *File       `json:"file"`
	Flag          string      `json:"flag"`
	Text          string      `json:"-"` // Known as "message", in a message or request
	Message       *Message    `json:"-"`
}

// UpdatesChannel is the channel for getting updates.
type UpdatesChannel <-chan Update

type File struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	BusID int64  `json:"busid"`
}

// User is a user on QQ.
type User struct {
	ID       int    `json:"user_id"`
	NickName string `json:"nickname"`
	Sex      string `json:"sex"` // "male"、"female"、"unknown"
	Age      int    `json:"age"`
	Area     string `json:"area"`
	// Group member
	Card                string `json:"card"`
	CardChangeable      bool   `json:"card_changeable"`
	Title               string `json:"title"`
	TitleExpireTimeUnix int64  `json:"title_expire_time"`
	Level               string `json:"level"`
	Role                string `json:"role"` // "owner"、"admin"、"member"
	Unfriendly          bool   `json:"unfriendly"`
	JoinTimeUnix        int64  `json:"join_time"`
	LastSentTimeUnix    int64  `json:"last_sent_time"`
	AnonymousName       string `json:"anonymous"`
	AnonymousFlag       string `json:"anonymous_flag"` // Anonymous ID
}

// String displays a simple text version of a user.
//
// It is normally a user's card, but falls back to a nickname as available.
func (u *User) String() string {
	p := ""
	if u.Title != "" {
		p = "[" + u.Title + "]"
	}
	return p + u.Name()
}

func (u *User) Name() string {
	if u.Card != "" {
		return u.Card
	}
	return u.NickName
}

// Chat contains information about the place a message was sent.
type Chat struct {
	ID      int64  `json:"id"`
	Type    string `json:"type"`     // "private"、"group"、"discuss"
	SubType string `json:"sub_type"` // (only when Type is "private") "friend"、"group"、"discuss"、"other"
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
	*cqcode.Message `json:"message"`
	MessageID       int    `json:"message_id"`
	From            *User  `json:"from"`
	Chat            *Chat  `json:"chat"`
	Text            string `json:"text"`
	SubType         string `json:"sub_type"` // (only when Chat.Type is "group") "normal"、"anonymous"、"notice"
	Font            int    `json:"font"`
}

func (m Message) IsAnonymous() bool {
	return m.SubType == "anonymous"
}

func (m Message) IsNotice() bool {
	return m.SubType == "notice"
}
