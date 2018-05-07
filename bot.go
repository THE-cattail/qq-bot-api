package qqbotapi

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type BotAPI struct {
	Token  string `json:"token"`
	Debug  bool   `json:"debug"`
	Buffer int    `json:"buffer"`

	Self         User         `json:"-"`
	Client       *http.Client `json:"-"`
	APIEndpoint  string       `json:"-"`
	PollEndpoint string       `json:"-"`
}

func NewBotAPI(token string, apiEndPoint string, pollEndpoint string) (*BotAPI, error) {
	return NewBotAPIWithClient(token, apiEndPoint, pollEndpoint, &http.Client{})
}

func NewBotAPIWithClient(token string, apiEndPoint string, pollEndpoint string, client *http.Client) (*BotAPI, error) {
	bot := &BotAPI{
		Token:        token,
		Client:       client,
		Buffer:       100,
		APIEndpoint:  apiEndPoint,
		PollEndpoint: pollEndpoint,
	}

	self, err := bot.GetMe()
	if err != nil {
		return nil, err
	}

	bot.Self = self

	return bot, nil
}

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

		if len(pollResp.Events) == 0 {
			return APIResponse{}, err
		}
		result := strings.Replace(string(pollResp.Events[0].Data), "\\\"", "\"", -1)
		apiResp := APIResponse{
			Result: json.RawMessage(result[1 : len(result)-1]),
		}
		if pollResp.Events[0].Error == "" {
			apiResp.Ok = "ok"
		} else {
			apiResp.Ok = pollResp.Events[0].Error
		}

		if apiResp.Ok != "ok" {
			parameters := ResponseParameters{}
			return apiResp, Error{apiResp.Ok, parameters}
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

	if apiResp.Ok != "ok" {
		parameters := ResponseParameters{}
		return apiResp, Error{apiResp.Ok, parameters}
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
	json.Unmarshal(resp.Result, &message)

	bot.debugLog(endpoint, params, message)

	return message, nil
}

func (bot *BotAPI) GetMe() (User, error) {
	resp, err := bot.MakeRequest("get_login_info", nil)
	if err != nil {
		return User{}, err
	}

	var user User
	json.Unmarshal(resp.Result, &user)
	user.UserName = strconv.Itoa(user.ID)

	bot.debugLog("getMe", nil, user)

	return user, nil
}

func (bot *BotAPI) IsMessageToMe(message Message) bool {
	return strings.Contains(message.Text, "@"+bot.Self.Nickname)
}

func (bot *BotAPI) Send(c Chattable) (Message, error) {
	switch c.(type) {
	default:
		return bot.sendChattable(c)
	}
}

func (bot *BotAPI) debugLog(context string, v url.Values, message interface{}) {
	if bot.Debug {
		log.Printf("%s req : %+v\n", context, v)
		log.Printf("%s resp: %+v\n", context, message)
	}
}

func (bot *BotAPI) sendChattable(config Chattable) (Message, error) {
	v, err := config.values()
	if err != nil {
		return Message{}, err
	}

	message, err := bot.makeMessageRequest(config.method(), v)

	if err != nil {
		return Message{}, err
	}

	return message, nil
}

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

	var updatejson UpdateJson
	json.Unmarshal(resp.Result, &updatejson)

	var update Update
	if updatejson.PostType == "message" {
		chat := &Chat{
			Type: updatejson.MessageType,
		}
		if updatejson.MessageType == "private" {
			chat.ID = int64(updatejson.UserID)
		} else if updatejson.MessageType == "group" {
			chat.ID = int64(updatejson.GroupID)
		} else {
			chat.ID = int64(updatejson.DiscussID)
		}
		v := url.Values{}
		v.Add("user_id", strconv.Itoa(updatejson.UserID))
		resp, err := bot.MakeRequest("get_stranger_info", v)
		nickname := ""
		if err != nil {
			log.Println(err)
		} else {
			var user User
			json.Unmarshal(resp.Result, &user)
			nickname = user.Nickname
		}
		update = Update{
			UpdateID: updatejson.MessageID,
			Message: &Message{
				From: &User{
					ID:       updatejson.UserID,
					UserName: strconv.Itoa(updatejson.UserID),
					Nickname: nickname,
					Cardname: "",
				},
				Text: updatejson.Message,
				Chat: chat,
			},
		}
	}
	var updates []Update
	updates = append(updates, update)

	bot.debugLog("get_updates", v, updates)

	return updates, nil
}

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
