package qqbotapi

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"
)

// APIResponse is a response from the Telegram API with the result
// stored raw.
type APIResponse struct {
	Ok          string              `json:"status"`
	Result      json.RawMessage     `json:"data"`
	ErrorCode   int                 `json:"retcode"`
	Description string              `json:"description"`
	Parameters  *ResponseParameters `json:"parameters"`
}

type PollEventResponse struct {
	Category string          `json:"category"`
	Data     json.RawMessage `json:"data"`
	Error    string          `json:"error"`
	Timeout  string          `json:"timeout"`
}

type PollResponse struct {
	Events []PollEventResponse `json:"events"`
}

// ResponseParameters are various errors that can be returned in APIResponse.
type ResponseParameters struct {
	MigrateToChatID int64 `json:"migrate_to_chat_id"` // optional
	RetryAfter      int   `json:"retry_after"`        // optional
}

type UpdateJson struct {
	PostType      string `json:"post_type"`
	MessageType   string `json:"message_type"`
	SubType       string `json:"sub_type"`
	MessageID     int    `json:"message_id"`
	GroupID       int    `json:"group_id"`
	DiscussID     int    `json:"discuss_id"`
	UserID        int    `json:"user_id"`
	Anonymous     string `json:"anonymous"`
	AnonymousFlag string `json:"anonymous_flag"`
	Message       string `json:"message"`
}

// Update is an update response, from GetUpdates.
type Update struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message"`
}

// UpdatesChannel is the channel for getting updates.
type UpdatesChannel <-chan Update

// Clear discards all unprocessed incoming updates.
func (ch UpdatesChannel) Clear() {
	for len(ch) != 0 {
		<-ch
	}
}

// User is a user on Telegram.
type User struct {
	ID       int    `json:"user_id"`
	UserName string `json:"-"`
	Nickname string `json:"nickname"`
	Cardname string `json:"cardname"`
}

// String displays a simple text version of a user.
//
// It is normally a user's username, but falls back to a first/last
// name as available.
func (u *User) String() string {
	if u.Cardname != "" {
		return u.Cardname
	}
	return u.Nickname
}

// GroupChat is a group chat.
type GroupChat struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// ChatPhoto represents a chat photo.
type ChatPhoto struct {
	SmallFileID string `json:"small_file_id"`
	BigFileID   string `json:"big_file_id"`
}

// Chat contains information about the place a message was sent.
type Chat struct {
	ID                  int64      `json:"id"`
	Type                string     `json:"type"`
	Title               string     `json:"title"`                          // optional
	UserName            string     `json:"username"`                       // optional
	FirstName           string     `json:"first_name"`                     // optional
	LastName            string     `json:"last_name"`                      // optional
	AllMembersAreAdmins bool       `json:"all_members_are_administrators"` // optional
	Photo               *ChatPhoto `json:"photo"`
	Description         string     `json:"description,omitempty"` // optional
	InviteLink          string     `json:"invite_link,omitempty"` // optional
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

// ChatConfig returns a ChatConfig struct for chat related methods.
func (c Chat) ChatConfig() ChatConfig {
	return ChatConfig{ChatID: c.ID}
}

// Message is returned by almost every request, and contains data about
// almost anything.
type Message struct {
	MessageID             int              `json:"message_id"`
	From                  *User            `json:"from"` // optional
	Date                  int              `json:"date"`
	Chat                  *Chat            `json:"chat"`
	ForwardFrom           *User            `json:"forward_from"`            // optional
	ForwardFromChat       *Chat            `json:"forward_from_chat"`       // optional
	ForwardFromMessageID  int              `json:"forward_from_message_id"` // optional
	ForwardDate           int              `json:"forward_date"`            // optional
	ReplyToMessage        *Message         `json:"reply_to_message"`        // optional
	EditDate              int              `json:"edit_date"`               // optional
	Text                  string           `json:"text"`                    // optional
	Entities              *[]MessageEntity `json:"entities"`                // optional
	Caption               string           `json:"caption"`                 // optional
	NewChatMembers        *[]User          `json:"new_chat_members"`        // optional
	LeftChatMember        *User            `json:"left_chat_member"`        // optional
	NewChatTitle          string           `json:"new_chat_title"`          // optional
	DeleteChatPhoto       bool             `json:"delete_chat_photo"`       // optional
	GroupChatCreated      bool             `json:"group_chat_created"`      // optional
	SuperGroupChatCreated bool             `json:"supergroup_chat_created"` // optional
	ChannelChatCreated    bool             `json:"channel_chat_created"`    // optional
	MigrateToChatID       int64            `json:"migrate_to_chat_id"`      // optional
	MigrateFromChatID     int64            `json:"migrate_from_chat_id"`    // optional
	PinnedMessage         *Message         `json:"pinned_message"`          // optional
}

// Time converts the message timestamp into a Time.
func (m *Message) Time() time.Time {
	return time.Unix(int64(m.Date), 0)
}

// IsCommand returns true if message starts with a "bot_command" entity.
func (m *Message) IsCommand() bool {
	if m.Entities == nil || len(*m.Entities) == 0 {
		return false
	}

	entity := (*m.Entities)[0]
	return entity.Offset == 0 && entity.Type == "bot_command"
}

// Command checks if the message was a command and if it was, returns the
// command. If the Message was not a command, it returns an empty string.
//
// If the command contains the at name syntax, it is removed. Use
// CommandWithAt() if you do not want that.
func (m *Message) Command() string {
	command := m.CommandWithAt()

	if i := strings.Index(command, "@"); i != -1 {
		command = command[:i]
	}

	return command
}

// CommandWithAt checks if the message was a command and if it was, returns the
// command. If the Message was not a command, it returns an empty string.
//
// If the command contains the at name syntax, it is not removed. Use Command()
// if you want that.
func (m *Message) CommandWithAt() string {
	if !m.IsCommand() {
		return ""
	}

	// IsCommand() checks that the message begins with a bot_command entity
	entity := (*m.Entities)[0]
	return m.Text[1:entity.Length]
}

// CommandArguments checks if the message was a command and if it was,
// returns all text after the command name. If the Message was not a
// command, it returns an empty string.
//
// Note: The first character after the command name is omitted:
// - "/foo bar baz" yields "bar baz", not " bar baz"
// - "/foo-bar baz" yields "bar baz", too
// Even though the latter is not a command conforming to the spec, the API
// marks "/foo" as command entity.
func (m *Message) CommandArguments() string {
	if !m.IsCommand() {
		return ""
	}

	// IsCommand() checks that the message begins with a bot_command entity
	entity := (*m.Entities)[0]
	if len(m.Text) == entity.Length {
		return "" // The command makes up the whole message
	}

	return m.Text[entity.Length+1:]
}

// MessageEntity contains information about data in a Message.
type MessageEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	URL    string `json:"url"`  // optional
	User   *User  `json:"user"` // optional
}

// ParseURL attempts to parse a URL contained within a MessageEntity.
func (entity MessageEntity) ParseURL() (*url.URL, error) {
	if entity.URL == "" {
		return nil, errors.New(ErrBadURL)
	}

	return url.Parse(entity.URL)
}

// ChatMember is information about a member in a chat.
type ChatMember struct {
	User                  *User  `json:"user"`
	Status                string `json:"status"`
	UntilDate             int64  `json:"until_date,omitempty"`                // optional
	CanBeEdited           bool   `json:"can_be_edited,omitempty"`             // optional
	CanChangeInfo         bool   `json:"can_change_info,omitempty"`           // optional
	CanPostMessages       bool   `json:"can_post_messages,omitempty"`         // optional
	CanEditMessages       bool   `json:"can_edit_messages,omitempty"`         // optional
	CanDeleteMessages     bool   `json:"can_delete_messages,omitempty"`       // optional
	CanInviteUsers        bool   `json:"can_invite_users,omitempty"`          // optional
	CanRestrictMembers    bool   `json:"can_restrict_members,omitempty"`      // optional
	CanPinMessages        bool   `json:"can_pin_messages,omitempty"`          // optional
	CanPromoteMembers     bool   `json:"can_promote_members,omitempty"`       // optional
	CanSendMessages       bool   `json:"can_send_messages,omitempty"`         // optional
	CanSendMediaMessages  bool   `json:"can_send_media_messages,omitempty"`   // optional
	CanSendOtherMessages  bool   `json:"can_send_other_messages,omitempty"`   // optional
	CanAddWebPagePreviews bool   `json:"can_add_web_page_previews,omitempty"` // optional
}

// IsCreator returns if the ChatMember was the creator of the chat.
func (chat ChatMember) IsCreator() bool { return chat.Status == "creator" }

// IsAdministrator returns if the ChatMember is a chat administrator.
func (chat ChatMember) IsAdministrator() bool { return chat.Status == "administrator" }

// IsMember returns if the ChatMember is a current member of the chat.
func (chat ChatMember) IsMember() bool { return chat.Status == "member" }

// HasLeft returns if the ChatMember left the chat.
func (chat ChatMember) HasLeft() bool { return chat.Status == "left" }

// WasKicked returns if the ChatMember was kicked from the chat.
func (chat ChatMember) WasKicked() bool { return chat.Status == "kicked" }

// Error is an error containing extra information returned by the Telegram API.
type Error struct {
	Message string
	ResponseParameters
}

func (e Error) Error() string {
	return e.Message
}
