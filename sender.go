package qqbotapi

import (
	"github.com/catsworld/qq-bot-api/cqcode"
	"net/url"
)

type FlatSender struct {
	bot      *BotAPI
	ChatID   int64
	ChatType string
	cache    cqcode.Message
	Result   *Message
	Err      error
}

type Sender struct {
	*FlatSender
}

func clone(sender *FlatSender) *FlatSender {
	newCache := make(cqcode.Message, 0)
	copy(newCache, sender.cache)
	return &FlatSender{
		bot:      sender.bot,
		ChatID:   sender.ChatID,
		ChatType: sender.ChatType,
		cache:    newCache,
		Result:   nil,
		Err:      nil,
	}
}

func NewSender(bot *BotAPI, chatID int64, chatType string) *Sender {
	return &Sender{
		FlatSender: &FlatSender{
			bot:      bot,
			ChatID:   chatID,
			ChatType: chatType,
			cache:    make(cqcode.Message, 0),
			Result:   nil,
			Err:      nil,
		},
	}
}

func (sender *FlatSender) Send() *Sender {
	msg, err := sender.bot.SendMessage(sender.ChatID, sender.ChatType, sender.cache)
	return &Sender{
		FlatSender: &FlatSender{
			bot:      sender.bot,
			ChatID:   sender.ChatID,
			ChatType: sender.ChatType,
			cache:    make(cqcode.Message, 0),
			Result:   &msg,
			Err:      err,
		},
	}
}

func (sender *FlatSender) ImageBase64(file interface{}) *FlatSender {
	n := clone(sender)
	img, err := NewImageBase64(file)
	if err == nil {
		n.cache = append(n.cache, img)
	}
	return n
}

func (sender *Sender) RecordBase64(file interface{}, magic bool) *Sender {
	n := clone(sender.FlatSender)
	rec, err := NewRecordBase64(file)
	if err == nil {
		rec.Magic = magic
		n.cache = append(n.cache, rec)
	}
	return n.Send()
}

// This method is deprecated and will get removed, see #11.
// Please use ImageWeb instead.
func (sender *FlatSender) ImageLocal(file string) *FlatSender {
	n := clone(sender)
	img := NewImageLocal(file)
	n.cache = append(n.cache, img)
	return n
}

// This method is deprecated and will get removed, see #11.
// Please use RecordWeb instead.
func (sender *Sender) RecordLocal(file string, magic bool) *Sender {
	n := clone(sender.FlatSender)
	rec := NewRecordLocal(file)
	rec.Magic = magic
	n.cache = append(n.cache, rec)
	return n.Send()
}

func (sender *FlatSender) ImageWeb(url *url.URL) *FlatSender {
	n := clone(sender)
	img := NewImageWeb(url)
	n.cache = append(n.cache, img)
	return n
}

func (sender *Sender) RecordWeb(url *url.URL, magic bool) *Sender {
	n := clone(sender.FlatSender)
	rec := NewRecordWeb(url)
	rec.Magic = magic
	n.cache = append(n.cache, rec)
	return n.Send()
}

func (sender *FlatSender) Text(text string) *FlatSender {
	n := clone(sender)
	t := cqcode.Text{
		Text: text,
	}
	n.cache = append(n.cache, &t)
	return n
}

func (sender *FlatSender) NewLine() *FlatSender {
	n := clone(sender)
	t := cqcode.Text{
		Text: "\n",
	}
	n.cache = append(n.cache, &t)
	return n
}

func (sender *FlatSender) At(QQ string) *FlatSender {
	n := clone(sender)
	t := cqcode.At{
		QQ: QQ,
	}
	n.cache = append(n.cache, &t)
	return n
}

func (sender *FlatSender) Face(faceID int) *FlatSender {
	n := clone(sender)
	t := cqcode.Face{
		FaceID: faceID,
	}
	n.cache = append(n.cache, &t)
	return n
}

func (sender *FlatSender) FaceByName(faceName string) *FlatSender {
	n := clone(sender)
	t, err := cqcode.NewFaceFromName(faceName)
	if err == nil {
		n.cache = append(n.cache, t)
	}
	return n
}

func (sender *FlatSender) Emoji(emojiID int) *FlatSender {
	n := clone(sender)
	t := cqcode.Emoji{
		EmojiID: emojiID,
	}
	n.cache = append(n.cache, &t)
	return n
}

func (sender *Sender) Bface(bfaceID int) *Sender {
	n := clone(sender.FlatSender)
	t := cqcode.Bface{
		BfaceID: bfaceID,
	}
	n.cache = append(n.cache, &t)
	return n.Send()
}

func (sender *FlatSender) Sface(sfaceID int) *FlatSender {
	n := clone(sender)
	t := cqcode.Sface{
		SfaceID: sfaceID,
	}
	n.cache = append(n.cache, &t)
	return n
}

func (sender *Sender) Rps() *Sender {
	n := clone(sender.FlatSender)
	t := cqcode.Rps{
		Type: 0,
	}
	n.cache = append(n.cache, &t)
	return n.Send()
}

func (sender *Sender) Dice() *Sender {
	n := clone(sender.FlatSender)
	t := cqcode.Dice{
		Type: 0,
	}
	n.cache = append(n.cache, &t)
	return n.Send()
}

func (sender *Sender) Shake(emojiID int) *Sender {
	n := clone(sender.FlatSender)
	t := cqcode.Emoji{
		EmojiID: emojiID,
	}
	n.cache = append(n.cache, &t)
	return n.Send()
}

func (sender *Sender) Music(music cqcode.Music) *Sender {
	n := clone(sender.FlatSender)
	n.cache = append(n.cache, &music)
	return n.Send()
}

func (sender *Sender) Share(share cqcode.Share) *Sender {
	n := clone(sender.FlatSender)
	n.cache = append(n.cache, &share)
	return n.Send()
}

func (sender *Sender) Location(loc cqcode.Location) *Sender {
	n := clone(sender.FlatSender)
	n.cache = append(n.cache, &loc)
	return n.Send()
}

func (sender *Sender) Show(id int) *Sender {
	n := clone(sender.FlatSender)
	t := cqcode.Show{
		ID: id,
	}
	n.cache = append(n.cache, &t)
	return n.Send()
}

func (sender *Sender) Sign(sign cqcode.Sign) *Sender {
	n := clone(sender.FlatSender)
	n.cache = append(n.cache, &sign)
	return n.Send()
}
