package qqbotapi

import (
	"net/url"
	"strconv"
	"time"
)

// Chattable is any config type that can be sent.
type Chattable interface {
	values() (url.Values, error)
	method() string
}

// BaseChat is base type for all chat config types.
type BaseChat struct {
	ChatID   int64 // required
	ChatType string
}

// values returns url.Values representation of BaseChat.
func (chat *BaseChat) values() (url.Values, error) {
	v := url.Values{}
	v.Add("message_type", chat.ChatType)
	switch chat.ChatType {
	case "private":
		v.Add("user_id", strconv.FormatInt(chat.ChatID, 10))
	case "group":
		v.Add("group_id", strconv.FormatInt(chat.ChatID, 10))
	case "discuss":
		v.Add("discuss_id", strconv.FormatInt(chat.ChatID, 10))
	}

	return v, nil
}

// MessageConfig contains information about a SendMessage request.
type MessageConfig struct {
	BaseChat
	Text       string
	AutoEscape bool
}

// values returns a url.Values representation of MessageConfig.
func (config MessageConfig) values() (url.Values, error) {
	v, err := config.BaseChat.values()
	if err != nil {
		return v, err
	}

	v.Add("message", config.Text)
	v.Add("auto_escape", strconv.FormatBool(config.AutoEscape))

	return v, nil
}

// method returns CQ HTTP API method name for sending message.
func (config MessageConfig) method() string {
	return "send_msg"
}

// DeleteMessageConfig contains information of a message in a chat to delete.
type DeleteMessageConfig struct {
	MessageID int64
}

// method returns CQ HTTP API method name for deleting message.
func (config DeleteMessageConfig) method() string {
	return "delete_msg"
}

// values returns url.Values representation of DeleteMessageConfig.
func (config DeleteMessageConfig) values() (url.Values, error) {
	v := url.Values{}

	v.Add("message_id", strconv.FormatInt(config.MessageID, 10))

	return v, nil
}

// LikeConfig contains information of a like (displayed on personal profile page) to send.
type LikeConfig struct {
	UserID int64
	Times  int
}

// method returns CQ HTTP API method name for sending like.
func (config LikeConfig) method() string {
	return "send_like"
}

// values returns url.Values representation of LikeConfig.
func (config LikeConfig) values() (url.Values, error) {
	v := url.Values{}

	v.Add("user_id", strconv.FormatInt(config.UserID, 10))
	v.Add("times", strconv.Itoa(config.Times))

	return v, nil
}

// ChatMemberConfig contains information about a user in a chat for use
// with administrative functions such as kicking or unbanning a user.
type ChatMemberConfig struct {
	GroupID       int64
	UserID        int64
	AnonymousFlag string
}

// values returns url.Values representation of ChatMemberConfig.
func (config ChatMemberConfig) values() (url.Values, error) {
	v := url.Values{}

	v.Add("group_id", strconv.FormatInt(config.GroupID, 10))
	v.Add("user_id", strconv.FormatInt(config.UserID, 10))
	v.Add("flag", config.AnonymousFlag)

	return v, nil
}

// KickChatMemberConfig contains extra fields to kick user.
type KickChatMemberConfig struct {
	ChatMemberConfig
	RejectAddRequest bool
}

// method returns CQ HTTP API method name for kicking user.
func (config KickChatMemberConfig) method() string {
	return "set_group_kick"
}

// values returns url.Values representation of KickChatMemberConfig.
func (config KickChatMemberConfig) values() (url.Values, error) {
	v, err := config.ChatMemberConfig.values()
	if err != nil {
		return v, err
	}

	v.Add("reject_add_request", strconv.FormatBool(config.RejectAddRequest))

	return v, nil
}

// RestrictChatMemberConfig contains fields to restrict members of chat.
type RestrictChatMemberConfig struct {
	ChatMemberConfig
	Duration time.Duration
}

// method returns CQ HTTP API method name for restricting user.
func (config RestrictChatMemberConfig) method() string {
	if config.AnonymousFlag != "" {
		return "set_group_anonymous_ban"
	}
	return "set_group_ban"
}

// values returns url.Values representation of RestrictChatMemberConfig.
func (config RestrictChatMemberConfig) values() (url.Values, error) {
	v, err := config.ChatMemberConfig.values()
	if err != nil {
		return v, err
	}

	v.Add("duration", strconv.FormatFloat(config.Duration.Seconds(), 'f', -1, 64))

	return v, nil
}

// PromoteChatMemberConfig contains fields to promote members of chat.
type PromoteChatMemberConfig struct {
	ChatMemberConfig
	Enable bool
}

// method returns CQ HTTP API method name for promoting user.
func (config PromoteChatMemberConfig) method() string {
	return "set_group_admin"
}

// values returns url.Values representation of PromoteChatMemberConfig.
func (config PromoteChatMemberConfig) values() (url.Values, error) {
	v, err := config.ChatMemberConfig.values()
	if err != nil {
		return v, err
	}

	v.Add("enable", strconv.FormatBool(config.Enable))

	return v, nil
}

// SetChatMemberCardConfig contains fields to set members's 群名片.
type SetChatMemberCardConfig struct {
	ChatMemberConfig
	Card string
}

// method returns CQ HTTP API method name for setting card.
func (config SetChatMemberCardConfig) method() string {
	return "set_group_card"
}

// values returns url.Values representation of SetChatMemberCardConfig.
func (config SetChatMemberCardConfig) values() (url.Values, error) {
	v, err := config.ChatMemberConfig.values()
	if err != nil {
		return v, err
	}

	v.Add("card", config.Card)

	return v, nil
}

// SetChatMemberTitleConfig contains fields to set members's 专属头衔.
type SetChatMemberTitleConfig struct {
	ChatMemberConfig
	SpecialTitle string
	Duration     time.Duration
}

// method returns CQ HTTP API method name for setting title.
func (config SetChatMemberTitleConfig) method() string {
	return "set_group_card"
}

// values returns url.Values representation of SetChatMemberTitleConfig.
func (config SetChatMemberTitleConfig) values() (url.Values, error) {
	v, err := config.ChatMemberConfig.values()
	if err != nil {
		return v, err
	}

	v.Add("special_title", config.SpecialTitle)
	v.Add("duration", strconv.FormatFloat(config.Duration.Seconds(), 'f', -1, 64))

	return v, nil
}

// GroupControlConfig contains fields as a configuration of a group.
type GroupControlConfig struct {
	GroupID int64
	Enable  bool
}

// values returns url.Values representation of GroupControlConfig.
func (config GroupControlConfig) values() (url.Values, error) {
	v := url.Values{}

	v.Add("group_id", strconv.FormatInt(config.GroupID, 10))
	v.Add("enable", strconv.FormatBool(config.Enable))

	return v, nil
}

// RestrictAllChatMembersConfig contains fields to restrict all chat members.
type RestrictAllChatMembersConfig struct {
	GroupControlConfig
}

// method returns CQ HTTP API method name for restricting all members.
func (config RestrictAllChatMembersConfig) method() string {
	return "set_group_whole_ban"
}

// EnableAnonymousChatConfig contains fields to enable anonymous chat.
type EnableAnonymousChatConfig struct {
	GroupControlConfig
}

// method returns CQ HTTP API method name for sending Message.
func (config EnableAnonymousChatConfig) method() string {
	return "set_group_anonymous"
}

// LeaveChatConfig contains fields to leave a chat.
type LeaveChatConfig struct {
	BaseChat
	IsDismiss bool
}

// method returns CQ HTTP API method name for leaving chat.
func (config LeaveChatConfig) method() string {
	switch config.ChatType {
	case "discuss":
		return "set_discuss_leave"
	default:
		return "set_group_leave"
	}
}

// values returns url.Values representation of LeaveChatConfig.
func (config LeaveChatConfig) values() (url.Values, error) {
	v, err := config.BaseChat.values()
	if err != nil {
		return v, err
	}

	v.Add("is_dismiss", strconv.FormatBool(config.IsDismiss))

	return v, nil
}

// HandleRequestConfig contains fields to handle a request.
type HandleRequestConfig struct {
	RequestFlag string
	Approve     bool
}

// values returns url.Values representation of HandleRequestConfig.
func (config HandleRequestConfig) values() (url.Values, error) {
	v := url.Values{}

	v.Add("flag", config.RequestFlag)
	v.Add("approve", strconv.FormatBool(config.Approve))

	return v, nil
}

// HandleFriendRequestConfig contains fields to handle a friend request.
type HandleFriendRequestConfig struct {
	HandleRequestConfig
	Remark string
}

// method returns CQ HTTP API method name for handling friend requests.
func (config HandleFriendRequestConfig) method() string {
	return "set_friend_add_request"
}

// values returns url.Values representation of HandleFriendRequestConfig.
func (config HandleFriendRequestConfig) values() (url.Values, error) {
	v, err := config.HandleRequestConfig.values()
	if err != nil {
		return v, err
	}

	v.Add("remark", config.Remark)

	return v, nil
}

// HandleGroupRequestConfig contains fields to handle a group adding request.
type HandleGroupRequestConfig struct {
	HandleRequestConfig
	Type   string
	Reason string
}

// method returns CQ HTTP API method name for handling group adding requests.
func (config HandleGroupRequestConfig) method() string {
	return "set_group_add_request"
}

// values returns url.Values representation of HandleGroupRequestConfig.
func (config HandleGroupRequestConfig) values() (url.Values, error) {
	v, err := config.HandleRequestConfig.values()
	if err != nil {
		return v, err
	}

	v.Add("type", config.Type)
	v.Add("reason", config.Reason)

	return v, nil
}

// UpdateConfig contains information about a GetUpdates request.
type UpdateConfig struct {
	BaseUpdateConfig
	Offset  int
	Limit   int
	Timeout int
}

// WebhookConfig contains information about a webhook.
type WebhookConfig struct {
	BaseUpdateConfig
	Pattern string // the webhook endpoint
}

// BaseUpdateConfig contains information about loading updates.
type BaseUpdateConfig struct {
	PreloadUserInfo bool // if this is enabled, more information will be provided in Update.From
}
