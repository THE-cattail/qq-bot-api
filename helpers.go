package qqbotapi

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"

	"github.com/catsworld/qq-bot-api/cqcode"
)

// NewMessage creates a new Message.
//
// chatID is where to send it, message is the message.
func NewMessage(chatID int64, chatType string, message interface{}) MessageConfig {
	mc := MessageConfig{
		BaseChat: BaseChat{
			ChatID:   chatID,
			ChatType: chatType,
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

// NewWebhook registers a webhook.
func NewWebhook(pattern string) WebhookConfig {
	return WebhookConfig{
		BaseUpdateConfig: BaseUpdateConfig{
			PreloadUserInfo: false,
		},
		Pattern: pattern,
	}
}

const (
	cacheEnabled  = 1
	cacheDisabled = 0
)

// NetResource is a resource located in the Internet.
type NetResource struct {
	Cache int `cq:"cache"`
}

// EnableCache enables CQ HTTP's cache feature.
func (r *NetResource) EnableCache() {
	r.Cache = cacheEnabled
}

// DisableCache forces CQ HTTP download from the URL instead of using cache.
func (r *NetResource) DisableCache() {
	r.Cache = cacheDisabled
}

// NetImage is an image located in the Internet.
type NetImage struct {
	*cqcode.Image
	*NetResource
}

// NetRecord is a record located in the Internet.
type NetRecord struct {
	*cqcode.Record
	*NetResource
}

// NewImageBase64 formats an image in base64.
func NewImageBase64(file interface{}) (*cqcode.Image, error) {
	fileid, err := NewFileBase64(file)
	if err != nil {
		return &cqcode.Image{}, err
	}
	return &cqcode.Image{
		FileID: fileid,
	}, nil
}

// NewRecordBase64 formats a record in base64.
func NewRecordBase64(file interface{}) (*cqcode.Record, error) {
	fileid, err := NewFileBase64(file)
	if err != nil {
		return &cqcode.Record{}, err
	}
	return &cqcode.Record{
		FileID: fileid,
	}, nil
}

// NewFileBase64 formats a file into base64 format.
func NewFileBase64(file interface{}) (string, error) {
	switch f := file.(type) {
	case string:
		data, err := ioutil.ReadFile(f)
		if err != nil {
			return "", err
		}
		encodeString := base64.StdEncoding.EncodeToString(data)
		return "base64://" + encodeString, nil

	case []byte:

		encodeString := base64.StdEncoding.EncodeToString(f)
		return "base64://" + encodeString, nil

	case io.Reader:

		data, err := ioutil.ReadAll(f)
		if err != nil {
			return "", err
		}

		encodeString := base64.StdEncoding.EncodeToString(data)
		return "base64://" + encodeString, nil

	default:
		return "", errors.New("bad file type")
	}
}

// NewImageLocal formats an image with the file path,
// this requires CQ HTTP runs in the same host with your bot.
func NewImageLocal(file string) *cqcode.Image {
	return &cqcode.Image{
		FileID: NewFileLocal(file),
	}
}

// NewRecordLocal formats a record with the file path,
// this requires CQ HTTP runs in the same host with your bot.
func NewRecordLocal(file string) *cqcode.Record {
	return &cqcode.Record{
		FileID: NewFileLocal(file),
	}
}

// NewFileLocal formats a file with the file path, returning the string.
func NewFileLocal(file string) string {
	return "file://" + file
}

// NewImageWeb formats an image with the URL.
func NewImageWeb(url *url.URL) *NetImage {
	return &NetImage{
		Image: &cqcode.Image{
			FileID: url.String(),
		},
		NetResource: &NetResource{
			Cache: cacheEnabled,
		},
	}
}

// NewRecordWeb formats a record with the URL.
func NewRecordWeb(url *url.URL) *NetRecord {
	return &NetRecord{
		Record: &cqcode.Record{
			FileID: url.String(),
		},
		NetResource: &NetResource{
			Cache: cacheEnabled,
		},
	}
}
