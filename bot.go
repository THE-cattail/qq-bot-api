// Package qqbotapi has functions and types used for interacting with
// the Coolq HTTP API.
package qqbotapi

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/catsworld/qq-bot-api/cqcode"
	"github.com/mitchellh/mapstructure"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// BotAPI allows you to interact with the Coolq HTTP API.
type BotAPI struct {
	Token       string `json:"token"`
	Secret      string `json:"secret"`
	Debug       bool   `json:"debug"`
	Buffer      int    `json:"buffer"`
	APIEndpoint string `json:"api_endpoint"`

	Self   User         `json:"-"`
	Client *http.Client `json:"-"`
}

// NewBotAPI creates a new BotAPI instance.
//
// token: access_token, api: API Endpoint of Coolq-http, example: http://host:port.
// secret: the secret key of HMAC SHA1 signature of Coolq-http, won't be validated if left blank.
func NewBotAPI(token string, api string, secret string) (*BotAPI, error) {
	return NewBotAPIWithClient(token, api, secret, &http.Client{})
}

// NewBotAPIWithClient creates a new BotAPI instance
// and allows you to pass a http.Client.
//
// It requires a token, an API endpoint and a poll endpoint which you
// set in Coolq HTTP API.
func NewBotAPIWithClient(token string, api string, secret string, client *http.Client) (*BotAPI, error) {
	bot := &BotAPI{
		Token:       token,
		Client:      client,
		Buffer:      100,
		APIEndpoint: api,
		Secret:      secret,
	}

	self, err := bot.GetMe()
	if err != nil {
		return nil, err
	}

	bot.Self = self

	return bot, nil
}

// MakeRequest makes a request to a specific endpoint with our token.
func (bot *BotAPI) MakeRequest(endpoint string, params url.Values) (APIResponse, error) {

	method := fmt.Sprintf("%s/%s?access_token=%s", bot.APIEndpoint, endpoint, bot.Token)

	resp, err := bot.Client.PostForm(method, params)
	if err != nil {
		return APIResponse{}, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	bytes, err := bot.decodeAPIResponse(resp.Body, &apiResp)
	if err != nil {
		return apiResp, err
	}

	if bot.Debug {
		log.Printf("%s resp: %s", endpoint, bytes)
	}

	if apiResp.Status != "ok" {
		return apiResp, errors.New(apiResp.Status + " " + strconv.Itoa(apiResp.RetCode))
	}

	return apiResp, nil
}

func (bot *BotAPI) decodeAPIResponse(responseBody io.Reader, resp *APIResponse) (_ []byte, err error) {
	if !bot.Debug {
		dec := json.NewDecoder(responseBody)
		err = dec.Decode(resp)
		return
	}

	// if debug, read reponse body
	data, err := ioutil.ReadAll(responseBody)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, resp)
	if err != nil {
		return
	}

	return data, nil
}

func (bot *BotAPI) makeMessageRequest(endpoint string, params url.Values) (Message, error) {
	resp, err := bot.MakeRequest(endpoint, params)
	if err != nil {
		return Message{}, err
	}

	var message Message
	json.Unmarshal(resp.Data, &message)

	bot.debugLog(endpoint, params, message)

	return message, nil
}

// GetMe fetches the currently authenticated bot.
//
// This method is called upon creation to validate the token,
// and so you may get this data from BotAPI.Self without the need for
// another request.
func (bot *BotAPI) GetMe() (User, error) {
	resp, err := bot.MakeRequest("get_login_info", nil)
	if err != nil {
		return User{}, err
	}

	var user User
	json.Unmarshal(resp.Data, &user)

	bot.debugLog("GetMe", nil, user)

	return user, nil
}

// GetStrangerInfo fetches a stranger's user info.
func (bot *BotAPI) GetStrangerInfo(userID int64) (User, error) {
	v := url.Values{}
	v.Add("user_id", strconv.FormatInt(userID, 10))
	resp, err := bot.MakeRequest("get_stranger_info", v)
	if err != nil {
		return User{}, err
	}
	var user User
	json.Unmarshal(resp.Data, &user)

	bot.debugLog("GetStrangerInfo", nil, user)

	return user, nil
}

// GetGroupMemberInfo fetches a group member's user info.
//
// Using cache may result in not updating in time, but will be responded faster
func (bot *BotAPI) GetGroupMemberInfo(groupID int64, userID int64, noCache bool) (User, error) {
	v := url.Values{}
	v.Add("group_id", strconv.FormatInt(groupID, 10))
	v.Add("user_id", strconv.FormatInt(userID, 10))
	v.Add("no_cache", strconv.FormatBool(noCache))
	resp, err := bot.MakeRequest("get_group_member_info", v)
	if err != nil {
		return User{}, err
	}
	var user User
	json.Unmarshal(resp.Data, &user)

	bot.debugLog("GetGroupMemberInfo", nil, user)

	return user, nil
}

// GetGroupMemberList fetches a group all member's user info.
//
// This information might be not full or accurate enough.
func (bot *BotAPI) GetGroupMemberList(groupID int64) ([]User, error) {
	v := url.Values{}
	v.Add("group_id", strconv.FormatInt(groupID, 10))
	resp, err := bot.MakeRequest("get_group_member_list", v)
	if err != nil {
		return nil, err
	}
	users := make([]User, 0)
	json.Unmarshal(resp.Data, &users)

	bot.debugLog("GetGroupMemberInfo", nil, users)

	return users, nil
}

// GetGroupList fetches all groups
func (bot *BotAPI) GetGroupList() ([]Group, error) {
	v := url.Values{}
	resp, err := bot.MakeRequest("get_group_list", v)
	if err != nil {
		return nil, err
	}
	groups := make([]Group, 0)
	json.Unmarshal(resp.Data, &groups)

	bot.debugLog("GetGroupList", nil, groups)

	return groups, nil
}

// IsMessageToMe returns true if message directed to this bot.
//
// It requires the Message.
func (bot *BotAPI) IsMessageToMe(message Message) bool {
	for _, media := range *message.Message {
		at, ok := media.(*cqcode.At)
		if !ok {
			continue
		}
		if at.QQ == strconv.FormatInt(bot.Self.ID, 10) {
			return true
		}
	}
	return false
}

// Send will send a Chattable item to Coolq.
// The response will be regarded as Message, often with a MessageID in it.
//
// It requires the Chattable to send.
func (bot *BotAPI) Send(c Chattable) (Message, error) {
	v, err := c.values()
	if err != nil {
		return Message{}, err
	}

	message, err := bot.makeMessageRequest(c.method(), v)

	if err != nil {
		return Message{}, err
	}

	return message, nil
}

func (bot *BotAPI) debugLog(context string, message ...interface{}) {
	if bot.Debug {
		for i, v := range message {
			log.Printf("%s [%d]: %+v\n", context, i, v)
		}
	}
}

// Do will send a Chattable item to Coolq.
//
// It requires the Chattable to send.
func (bot *BotAPI) Do(c Chattable) (APIResponse, error) {
	v, err := c.values()
	if err != nil {
		return APIResponse{}, err
	}

	resp, err := bot.MakeRequest(c.method(), v)

	if err != nil {
		return APIResponse{}, err
	}

	return resp, nil
}

// ParseRawMessage parses message
func (update *Update) ParseRawMessage() {
	text, ok := update.RawMessage.(string)
	if update.PostType != "message" {
		update.Text = text
		return
	}
	messageSubType := "normal"
	chat := Chat{
		Type: update.MessageType,
	}
	if chat.IsPrivate() {
		chat.ID = int64(update.UserID)
		chat.SubType = update.SubType
	}
	if chat.IsGroup() {
		chat.ID = int64(update.GroupID)
		messageSubType = update.SubType
	}
	if chat.IsDiscuss() {
		chat.ID = int64(update.DiscussID)
	}
	message, _ := cqcode.ParseMessage(update.RawMessage)
	if !ok {
		text = message.CQString()
	}
	var user User
	if messageSubType == "anonymous" {
		anonymousName, ok := update.Anonymous.(string)
		if !ok {
			config := &mapstructure.DecoderConfig{
				Metadata:         nil,
				Result:           &user,
				WeaklyTypedInput: true,
				TagName:          "anonymous",
			}
			decoder, _ := mapstructure.NewDecoder(config)
			decoder.Decode(update.Anonymous)
		} else {
			user.AnonymousID = update.UserID
			user.AnonymousName = anonymousName
			user.AnonymousFlag = update.AnonymousFlag
		}
	}
	user.ID = update.UserID
	update.Message = &Message{
		Message:   &message,
		MessageID: update.MessageID,
		From:      &user,
		Chat:      &chat,
		Text:      text,
		SubType:   messageSubType,
	}
	if update.PostType == "event" {
		update.NoticeType = update.Event
	} else if update.PostType == "notice" {
		update.Event = update.NoticeType
	}
}

// PreloadUserInfo fills in the information in update.Message.From
func (bot *BotAPI) PreloadUserInfo(update *Update) {
	if update.Message == nil || update.Message.IsAnonymous() {
		return
	}
	var user User
	var err error
	if update.Message.Chat.Type == "group" {
		user, err = bot.GetGroupMemberInfo(update.GroupID, update.UserID, false)
		if err != nil {
			return
		}
	} else {
		user, err = bot.GetStrangerInfo(update.UserID)
		if err != nil {
			return
		}
	}

	update.Message.From = &user
}

// GetUpdates fetches updates over long polling.
// https://github.com/richardchien/coolq-http-api/issues/62
//
// Note that long polling is currently unsupported by coolq-http-api, thus this api
// might be changed in the future. It works with github.com/catsworld/cqhttp-longpoll-server at present.
//
// Offset, Limit, and Timeout are optional.
// To avoid stale items, set Offset to one higher than the previous item.
// Set Timeout to a large number to reduce requests so you can get updates
// instantly instead of having to wait between requests.
func (bot *BotAPI) GetUpdates(config UpdateConfig) ([]Update, error) {
	v := url.Values{}
	if config.Offset != 0 {
		v.Add("offset", strconv.Itoa(config.Offset))
	}
	if config.Limit > 0 {
		v.Add("limit", strconv.Itoa(config.Limit))
	}
	if config.Timeout > 0 {
		v.Add("timeout", strconv.Itoa(config.Timeout))
	}

	resp, err := bot.MakeRequest("get_updates", v)
	if err != nil {
		return []Update{}, err
	}

	var updates []Update
	json.Unmarshal(resp.Data, &updates)
	for i := range updates {
		updates[i].ParseRawMessage()
		if config.PreloadUserInfo {
			bot.PreloadUserInfo(&updates[i])
		}
	}

	bot.debugLog("getUpdates", v, updates)

	return updates, nil
}

// GetUpdatesChan starts and returns a channel that gets updates over long polling.
// https://github.com/richardchien/coolq-http-api/issues/62
//
// Note that long polling is currently unsupported by coolq-http-api, thus this api
// might be changed in the future. It works with github.com/catsworld/cqhttp-longpoll-server at present.
func (bot *BotAPI) GetUpdatesChan(config UpdateConfig) (UpdatesChannel, error) {
	ch := make(chan Update, bot.Buffer)

	go func() {
		for {
			updates, err := bot.GetUpdates(config)
			if err != nil {
				log.Println(err)
				log.Println("Failed to get updates, retrying in 3 seconds...")
				time.Sleep(time.Second * 3)

				continue
			}

			for _, update := range updates {
				ch <- update
			}
		}
	}()

	return ch, nil
}

// ListenForWebhook registers a http handler for a webhook and returns a channel that gets updates.
func (bot *BotAPI) ListenForWebhook(config WebhookConfig) UpdatesChannel {
	ch := make(chan Update, bot.Buffer)

	http.HandleFunc(config.Pattern, func(w http.ResponseWriter, r *http.Request) {
		bytes, _ := ioutil.ReadAll(r.Body)

		if bot.Secret != "" {
			mac := hmac.New(sha1.New, []byte(bot.Secret))
			mac.Write(bytes)
			expectedMac := r.Header.Get("X-Signature")[len("sha1="):]
			messageMac := hex.EncodeToString(mac.Sum(nil))
			if expectedMac != messageMac {
				bot.debugLog("ListenForWebhook HMAC", expectedMac, messageMac)
				return
			}
		}

		var update Update
		json.Unmarshal(bytes, &update)

		update.ParseRawMessage()
		if config.PreloadUserInfo {
			bot.PreloadUserInfo(&update)
		}

		bot.debugLog("ListenForWebhook", update)

		ch <- update

		w.WriteHeader(http.StatusNoContent)
	})

	return ch
}

// ListenForWebhookSync registers a http handler for a webhook.
//
// handler receives a update and returns a key-value dictionary.
func (bot *BotAPI) ListenForWebhookSync(config WebhookConfig, handler func(update Update) interface{}) {

	http.HandleFunc(config.Pattern, func(w http.ResponseWriter, r *http.Request) {
		bytes, _ := ioutil.ReadAll(r.Body)

		if bot.Secret != "" {
			mac := hmac.New(sha1.New, []byte(bot.Secret))
			mac.Write(bytes)
			expectedMac := r.Header.Get("X-Signature")[len("sha1="):]
			messageMac := hex.EncodeToString(mac.Sum(nil))
			if expectedMac != messageMac {
				bot.debugLog("ListenForWebhook HMAC", expectedMac, messageMac)
				return
			}
		}

		var update Update
		json.Unmarshal(bytes, &update)

		update.ParseRawMessage()
		if config.PreloadUserInfo {
			bot.PreloadUserInfo(&update)
		}

		bot.debugLog("ListenForWebhook", update)

		resp, _ := json.Marshal(handler(update))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	})
}

// SendMessage sends message to a chat.
func (bot *BotAPI) SendMessage(chatID int64, chatType string, message interface{}) (Message, error) {
	return bot.Send(NewMessage(chatID, chatType, message))
}

// DeleteMessage deletes a message in a chat.
func (bot *BotAPI) DeleteMessage(messageID int64) (APIResponse, error) {
	return bot.Do(DeleteMessageConfig{
		MessageID: messageID,
	})
}

// Like sends like (displayed in one's profile page) to a user.
func (bot *BotAPI) Like(userID int64, times int) (APIResponse, error) {
	return bot.Do(LikeConfig{
		UserID: userID,
		Times:  times,
	})
}

// KickChatMember kick a chat member in a group.
func (bot *BotAPI) KickChatMember(groupID int64, userID int64, rejectAddRequest bool) (APIResponse, error) {
	return bot.Do(KickChatMemberConfig{
		ChatMemberConfig: ChatMemberConfig{
			GroupID: groupID,
			UserID:  userID,
		},
		RejectAddRequest: rejectAddRequest,
	})
}

// RestrictChatMember bans a chat member from sending messages.
func (bot *BotAPI) RestrictChatMember(groupID int64, userID int64, duration time.Duration) (APIResponse, error) {
	return bot.Do(RestrictChatMemberConfig{
		ChatMemberConfig: ChatMemberConfig{
			GroupID: groupID,
			UserID:  userID,
		},
		Duration: duration,
	})
}

// RestrictAnonymousChatMember bans an anonymous chat member from sending messages.
func (bot *BotAPI) RestrictAnonymousChatMember(groupID int64, flag string, duration time.Duration) (APIResponse, error) {
	return bot.Do(RestrictChatMemberConfig{
		ChatMemberConfig: ChatMemberConfig{
			GroupID:       groupID,
			AnonymousFlag: flag,
		},
		Duration: duration,
	})
}

// RestrictAllChatMembers : By this enabled, only administrators in a group will be able to send messages.
func (bot *BotAPI) RestrictAllChatMembers(groupID int64, enable bool) (APIResponse, error) {
	return bot.Do(RestrictAllChatMembersConfig{
		GroupControlConfig: GroupControlConfig{
			GroupID: groupID,
			Enable:  enable,
		},
	})
}

// PromoteChatMember add admin rights to user.
func (bot *BotAPI) PromoteChatMember(groupID int64, userID int64, enable bool) (APIResponse, error) {
	return bot.Do(PromoteChatMemberConfig{
		ChatMemberConfig: ChatMemberConfig{
			GroupID: groupID,
			UserID:  userID,
		},
		Enable: enable,
	})
}

// EnableAnonymousChat : By this enabled, members in a group will be able to send messages with an anonymous identity.
func (bot *BotAPI) EnableAnonymousChat(groupID int64, enable bool) (APIResponse, error) {
	return bot.Do(EnableAnonymousChatConfig{
		GroupControlConfig: GroupControlConfig{
			GroupID: groupID,
			Enable:  enable,
		},
	})
}

// SetChatMemberCard sets a chat member's 群名片 in the group.
func (bot *BotAPI) SetChatMemberCard(groupID int64, userID int64, card string) (APIResponse, error) {
	return bot.Do(SetChatMemberCardConfig{
		ChatMemberConfig: ChatMemberConfig{
			GroupID: groupID,
			UserID:  userID,
		},
		Card: card,
	})
}

// SetChatMemberTitle sets a chat member's 专属头衔 in the group.
func (bot *BotAPI) SetChatMemberTitle(groupID int64, userID int64, title string, duration time.Duration) (APIResponse, error) {
	return bot.Do(SetChatMemberTitleConfig{
		ChatMemberConfig: ChatMemberConfig{
			GroupID: groupID,
			UserID:  userID,
		},
		SpecialTitle: title,
		Duration:     duration,
	})
}

// LeaveChat makes the bot leave the chat.
func (bot *BotAPI) LeaveChat(chatID int64, chatType string, dismiss bool) (APIResponse, error) {
	return bot.Do(LeaveChatConfig{
		BaseChat: BaseChat{
			ChatID:   chatID,
			ChatType: chatType,
		},
		IsDismiss: dismiss,
	})
}

// HandleFriendRequest handles a friend request.
//
// remark: 备注
func (bot *BotAPI) HandleFriendRequest(flag string, approve bool, remark string) (APIResponse, error) {
	return bot.Do(HandleFriendRequestConfig{
		HandleRequestConfig: HandleRequestConfig{
			RequestFlag: flag,
			Approve:     approve,
		},
		Remark: remark,
	})
}

// HandleGroupRequest handles a group adding request.
//
// typ: sub_type in Update
// reason: Reason if you rejects this request.
func (bot *BotAPI) HandleGroupRequest(flag string, typ string, approve bool, reason string) (APIResponse, error) {
	return bot.Do(HandleGroupRequestConfig{
		HandleRequestConfig: HandleRequestConfig{
			RequestFlag: flag,
			Approve:     approve,
		},
		Type:   typ,
		Reason: reason,
	})
}
