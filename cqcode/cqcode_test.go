package cqcode

import (
	"encoding/json"
	"testing"
)

func TestCQString(t *testing.T) {
	message := NewMessage()

	rec := Record{
		FileID: "/data/audio/[,]&",
		Magic:  false,
	}

	message = append(message, &rec)

	shake := Shake{}

	message = append(message, &shake)

	text := Text{
		Text: "[,]&",
	}

	message = append(message, &text)

	str := message.CQString()

	if str == "[CQ:record,file=/data/audio/&#91;&#44;&#93;&amp;,magic=false,url=][CQ:shake]&#91;,&#93;&amp;" {
		t.Log("Format CQ string passed")
	} else {
		t.Errorf("Format CQ string failed: %v", str)
	}
}

func TestParseCQCode(t *testing.T) {

	var text Text
	var face Face
	data := make([]interface{}, 0)

	cq1 := "&#91;he&#44;ym"

	err := ParseCQCode(cq1, &text)
	data = append(data, err)
	err = ParseCQCode(cq1, &face)
	data = append(data, err.Error())

	cq2 := "[CQ:face,id=14]"

	err = ParseCQCode(cq2, &text)
	data = append(data, err.Error())
	err = ParseCQCode(cq2, &face)
	data = append(data, err)

	data = append(data, face.FaceID)

	res, _ := json.Marshal(data)

	jsonstr := string(res)

	if jsonstr == `[null,"invalid cqcode","wrong media type",null,14]` {
		t.Log("Parse CQCode passed")
	} else {
		t.Errorf("Parse CQCode failed: %v", jsonstr)
	}

}

func TestMessageSegment_ParseMedia(t *testing.T) {

	seg := MessageSegment{
		Type: "text",
		Data: map[string]interface{}{
			"text": "test text message",
		},
	}

	var text Text
	seg.ParseMedia(&text)

	if text.Text == "test text message" {
		t.Log("Decode text passed")
	} else {
		t.Errorf("Decode text failed: %v", text.Text)
	}

}

func TestParseMessageFromString(t *testing.T) {

	str := "&#91;he&#44;ym[CQ:at,qq=123&#44;456][CQ:face,id=14] \nSee this awesome image, [CQ:image,file=1.jpg] Isn't it cool? [CQ:shake]\n"

	mes, err := ParseMessageFromString(str)

	if err != nil {
		t.Fatalf("Decode text failed: %v", err)
	}

	res, _ := json.Marshal(mes)

	jsonstr := string(res)

	if string(res) == `[{"Text":"[he,ym"},{"QQ":"123,456"},{"FaceID":14},{"Text":" \nSee this awesome image, "},{"FileID":"1.jpg","URL":""},{"Text":" Isn't it cool? "},{},{"Text":"\n"}]` {
		t.Log("Decode text passed")
	} else {
		t.Errorf("Decode text failed: %v", jsonstr)
	}

}

func TestMessage_Append(t *testing.T) {

	music := Music{
		Type:     "custom",
		ShareURL: "http://localhost:8080",
	}

	m := NewMessage()

	err := m.Append(&music)

	if err != nil {
		t.Fatalf("Decode text failed: %v", err)
	}

	res, _ := json.Marshal(m)

	jsonstr := string(res)

	if string(res) == `[{"Type":"custom","MusicID":"","ShareURL":"http://localhost:8080","AudioURL":"","Title":"","Content":"","Image":""}]` {
		t.Log("Append music passed")
	} else {
		t.Errorf("Append music failed: %v", jsonstr)
	}

}

func TestMessageSegment_CQString(t *testing.T) {

	shake := Shake{}

	seg, _ := NewMessageSegment(&shake)

	s := seg.CQString()

	text := Text{
		Text: "[,]&",
	}

	seg, _ = NewMessageSegment(&text)

	ts := seg.CQString()

	if s == "[CQ:shake]" && ts == "&#91;,&#93;&amp;" {
		t.Log("Format CQString passed")
	} else {
		t.Errorf("Format CQString failed: %v %v", s, ts)
	}

}

func TestCommand(t *testing.T) {

	m := NewMessage()

	text1 := Text{
		Text: "/",
	}

	m.Append(&text1)

	face := Face{
		FaceID: 170,
	}

	m.Append(&face)

	text2 := Text{
		Text: ` arg1 'a \'r 
g 2' "a \"r \\\"g 3\\" arg4
argemoji`,
	}

	m.Append(&text2)

	emoji := Emoji{
		EmojiID: 10086,
	}

	m.Append(&emoji)

	text3 := Text{
		Text: ` arg5`,
	}

	m.Append(&text3)

	music := Music{
		Content: "Alice\nLove\nBob",
	}

	m.Append(&music)

	StrictCommand = true

	if !m.IsCommand() {
		t.Error("Should be command")
	}

	cmd, args := m.Command()

	res, _ := json.Marshal(args)

	jsonstr := string(res)

	if cmd == "[CQ:face,id=170]" && jsonstr == `["arg1","a 'r \ng 2","a \"r \\\"g 3\\","arg4","argemoji[CQ:emoji,id=10086]","arg5[CQ:music,type=,id=,url=,audio=,title=,content=Alice\nLove\nBob,image=]"]` {
		t.Log("Good command")
	} else {
		t.Errorf("Parse command failed: %v", jsonstr)
	}

}
