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
	"strings"
	"time"

	"github.com/juzi5201314/cqhttp-go-sdk/cq"
)

// BotAPI allows you to interact with the Coolq HTTP API.
type BotAPI struct {
	Token  string `json:"token"`
	Debug  bool   `json:"debug"`
	Buffer int    `json:"buffer"`

	Self         User         `json:"-"`
	Client       *http.Client `json:"-"`
	APIEndpoint  string       `json:"-"`
	PollEndpoint string       `json:"-"`
}

// NewBotAPI creates a new BotAPI instance.
//
// It requires a token, an API endpoint and a poll endpoint which you
// set in Coolq HTTP API.
func NewBotAPI(token string, api string, poll string) (*BotAPI, error) {
	return NewBotAPIWithClient(token, api, poll, &http.Client{})
}

// NewBotAPIWithClient creates a new BotAPI instance
// and allows you to pass a http.Client.
//
// It requires a token, an API endpoint and a poll endpoint which you
// set in Coolq HTTP API.
func NewBotAPIWithClient(token string, api string, poll string, client *http.Client) (*BotAPI, error) {
	bot := &BotAPI{
		Token:        token,
		Client:       client,
		Buffer:       100,
		APIEndpoint:  api,
		PollEndpoint: poll,
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
	if endpoint == "get_updates" {
		method := fmt.Sprintf(bot.PollEndpoint, endpoint)

		resp, err := bot.Client.PostForm(method, params)
		if err != nil {
			return APIResponse{}, err
		}
		defer resp.Body.Close()

		var pollResp PollResponse
		bytes, err := bot.decodePollResponse(resp.Body, &pollResp)
		if err != nil {
			return APIResponse{}, err
		}

		if bot.Debug {
			log.Printf("%s resp: %s", endpoint, bytes)
		}

		if pollResp.Error != "" {
			return APIResponse{}, errors.New(pollResp.Error)
		}

		if len(pollResp.Events) == 0 {
			return APIResponse{}, errors.New("No poll events get")
		}

		var eventResp PollEvent
		json.Unmarshal(pollResp.Events[0], &eventResp)

		s := string(eventResp.Data)
		for i := 0; i < len(s); i++ {
			if s[i] == '\\' {
				s = s[:i] + s[i+1:]
			}
		}
		s = s[1 : len(s)-1]
		s = "[" + s + "]"
		apiResp := APIResponse{
			Status:  "ok",
			RetCode: 0,
			Data:    json.RawMessage(s),
		}

		return apiResp, nil
	}
	method := fmt.Sprintf(bot.APIEndpoint, endpoint, bot.Token)

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

func (bot *BotAPI) decodePollResponse(responseBody io.Reader, resp *PollResponse) (_ []byte, err error) {
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
	return strings.Contains(message.Text, cq.At(strconv.Itoa(bot.Self.ID)))
}

// Send will send a Chattable item to Coolq.
//
// It requires the Chattable to send.
func (bot *BotAPI) Send(c Chattable) (Message, error) {
	return bot.sendChattable(c)
}

func (bot *BotAPI) debugLog(context string, v url.Values, message interface{}) {
	if bot.Debug {
		log.Printf("%s req : %+v\n", context, v)
		log.Printf("%s resp: %+v\n", context, message)
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

// GetUpdates fetches updates.
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
	for i := 0; i < len(updates); i++ {
		chat := Chat{
			Type: updates[i].MessageType,
		}
		if chat.IsPrivate() {
			chat.ID = int64(updates[i].UserID)
		}
		if chat.IsGroup() {
			chat.ID = int64(updates[i].GroupID)
		}
		if chat.IsDiscuss() {
			chat.ID = int64(updates[i].DiscussID)
		}
		v := url.Values{}
		v.Add("user_id", strconv.Itoa(updates[i].UserID))
		resp, err := bot.MakeRequest("get_stranger_info", v)
		if err != nil {
			return []Update{}, err
		}
		if chat.Type == "group" {
			v := url.Values{}
			v.Add("group_id", strconv.Itoa(updates[i].GroupID))
			v.Add("user_id", strconv.Itoa(updates[i].UserID))
			resp, err = bot.MakeRequest("get_group_member_info", v)
			if err != nil {
				return []Update{}, err
			}
		}
		var user User
		json.Unmarshal(resp.Data, &user)
		updates[i].UpdateID = updates[i].MessageID
		updates[i].Message = &Message{
			MessageID: updates[i].MessageID,
			From:      &user,
			Chat:      &chat,
			Text:      updates[i].Text,
		}
	}

	bot.debugLog("getUpdates", v, updates)

	return updates, nil
}

// GetUpdatesChan starts and returns a channel for getting updates.
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
				if update.UpdateID >= config.Offset {
					config.Offset = update.UpdateID + 1
					ch <- update
				}
			}
		}
	}()

	return ch, nil
}
