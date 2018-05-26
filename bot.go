// Package qqbotapi has functions and types used for interacting with
// the Coolq HTTP API.
package qqbotapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/catsworld/qq-bot-api/cqcode"
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
// token: access_token, api: API Endpoint of Coolq-http, example: http://host:port
//
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
		return apiResp, errors.New(apiResp.Status + " " + strconv.FormatInt(apiResp.RetCode, 10))
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

	bot.debugLog("getMe", nil, user)

	return user, nil
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
		if at.QQ == strconv.Itoa(bot.Self.ID) {
			return true
		}
	}
	return false
}

// Send will send a Chattable item to Coolq.
//
// It requires the Chattable to send.
func (bot *BotAPI) Send(c Chattable) (Message, error) {
	return bot.sendChattable(c)
}

func (bot *BotAPI) debugLog(context string, message ...interface{}) {
	if bot.Debug {
		for i, v := range message {
			log.Printf("%s [%d]: %+v\n", context, i, v)
		}
	}
}

func (bot *BotAPI) sendChattable(config Chattable) (Message, error) {
	v, err := config.values()
	if bot.Debug && v.Get("message") != "" {
		t := "[Debug] " + v.Get("message")
		v.Set("message", t)
	}
	if err != nil {
		return Message{}, err
	}

	message, err := bot.makeMessageRequest(config.method(), v)

	if err != nil {
		return Message{}, err
	}

	return message, nil
}

// ParseRawMessage parses message
func (update Update) ParseRawMessage() {
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
	update.Message = &Message{
		Message:   &message,
		MessageID: update.MessageID,
		From: &User{
			ID:            update.UserID,
			AnonymousName: update.AnonymousName,
			AnonymousFlag: update.AnonymousFlag,
		},
		Chat:    &chat,
		Text:    text,
		SubType: messageSubType,
	}
}

// PreloadUserInfo fills in the information in update.Message.From
func (bot *BotAPI) PreloadUserInfo(update *Update) {
	if update.Message == nil || update.Message.IsAnonymous() {
		return
	}
	var resp APIResponse
	var err error
	if update.Message.Chat.Type == "group" {
		v := url.Values{}
		v.Add("group_id", strconv.Itoa(update.GroupID))
		v.Add("user_id", strconv.Itoa(update.UserID))
		resp, err = bot.MakeRequest("get_group_member_info", v)
		if err != nil {
			return
		}
	} else {
		v := url.Values{}
		v.Add("user_id", strconv.Itoa(update.UserID))
		resp, err = bot.MakeRequest("get_stranger_info", v)
		if err != nil {
			return
		}
	}
	var user User
	json.Unmarshal(resp.Data, &user)
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
	for _, update := range updates {
		update.ParseRawMessage()
		if config.PreloadUserInfo {
			bot.PreloadUserInfo(&update)
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

		var update Update
		json.Unmarshal(bytes, &update)

		update.ParseRawMessage()
		if config.PreloadUserInfo {
			bot.PreloadUserInfo(&update)
		}

		bot.debugLog("ListenForWebhook", update)

		ch <- update
	})

	return ch
}
