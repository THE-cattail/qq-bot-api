package qqbotapi

import (
	"encoding/json"
	"github.com/catsworld/qq-bot-api/cqcode"
	"net/url"
	"testing"
)

func TestNewImageWeb(t *testing.T) {
	u, _ := url.Parse("https://img.rikako.moe/i/D1D.jpg")
	img := NewImageWeb(u)
	img.DisableCache()
	str := cqcode.FormatCQCode(img)
	if str == "[CQ:image,file=https://img.rikako.moe/i/D1D.jpg,url=,cache=0]" {
		t.Log("TestNewImageWeb passed")
	} else {
		t.Errorf("TestNewImageWeb failed: %v", str)
	}
}

func TestNewMessage(t *testing.T) {
	image := cqcode.Image{
		FileID: "asjkdfs",
	}
	message := NewMessage(10000, "group", image)

	b, _ := json.Marshal(message)
	if string(b) == `{"ChatID":10000,"ChatType":"group","Text":"[CQ:image,file=asjkdfs,url=]","AutoEscape":false}` {
		t.Log("TestNewMessage passed")
	} else {
		t.Errorf("TestNewMessage failed: %v", string(b))
	}
}
