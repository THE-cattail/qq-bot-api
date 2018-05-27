package cqcode

import (
	"net/url"
	"testing"
)

func TestNewImageWeb(t *testing.T) {
	u, _ := url.Parse("https://img.rikako.moe/i/D1D.jpg")
	img := NewImageWeb(u)
	img.DisableCache()
	str := FormatCQCode(img)
	if str == "[CQ:image,file=https://img.rikako.moe/i/D1D.jpg,url=,cache=0]" {
		t.Log("TestNewImageWeb passed")
	} else {
		t.Errorf("TestNewImageWeb failed: %v", str)
	}
}
