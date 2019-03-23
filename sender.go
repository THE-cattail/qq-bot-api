package qqbotapi

import (
	"github.com/catsworld/qq-bot-api/cqcode"
	"net/url"
)

type FlatSender struct {
	bot *BotAPI
	chatID int64
	chatType string
	cache cqcode.Message
}

type Sender struct {
	*FlatSender
}

func NewSender(bot *BotAPI, chatID int64, chatType string) *Sender {
	return &Sender{
		FlatSender: &FlatSender{
			bot: bot,
			chatID:chatID,
			chatType:chatType,
			cache: make(cqcode.Message, 0),
		},
	}
}

func (sender *FlatSender) Send() (Message, error) {
	return sender.bot.SendMessage(sender.chatID, sender.chatType, sender.cache)
}

func (sender *FlatSender) ImageBase64(file interface{}) *FlatSender  {
	img, err := NewImageBase64(file)
	if err != nil {
		sender.cache = append(sender.cache, img)
	}
	return sender
}

func (sender *Sender) RecordBase64(file interface{}, magic bool) (Message, error)  {
	rec, err := NewRecordBase64(file)
	if err != nil {
		rec.Magic = magic
		sender.cache = append(sender.cache, rec)
	}
	return sender.Send()
}

// This method is deprecated and will get removed, see #11.
// Please use ImageWeb instead.
func (sender *FlatSender) ImageLocal(file string) *FlatSender  {
	img := NewImageLocal(file)
	sender.cache = append(sender.cache, img)
	return sender
}

// This method is deprecated and will get removed, see #11.
// Please use RecordWeb instead.
func (sender *Sender) RecordLocal(file string, magic bool) (Message, error)  {
	rec := NewRecordLocal(file)
	rec.Magic = magic
	sender.cache = append(sender.cache, rec)
	return sender.Send()
}

func (sender *FlatSender) ImageWeb(url *url.URL) *FlatSender  {
	img := NewImageWeb(url)
	sender.cache = append(sender.cache, img)
	return sender
}

func (sender *Sender) RecordWeb(url *url.URL, magic bool) (Message, error)  {
	rec := NewRecordWeb(url)
	rec.Magic = magic
	sender.cache = append(sender.cache, rec)
	return sender.Send()
}

func (sender *FlatSender) Text(text string) *FlatSender {
	t := cqcode.Text{
		Text: text,
	}
	sender.cache = append(sender.cache, &t)
	return sender
}

func (sender *FlatSender) NewLine() *FlatSender {
	t := cqcode.Text{
		Text: "\n",
	}
	sender.cache = append(sender.cache, &t)
	return sender
}

func (sender *FlatSender) At(QQ string) *FlatSender {
	t := cqcode.At{
		QQ: QQ,
	}
	sender.cache = append(sender.cache, &t)
	return sender
}

func (sender *FlatSender) Face(faceID int) *FlatSender {
	t := cqcode.Face{
		FaceID: faceID,
	}
	sender.cache = append(sender.cache, &t)
	return sender
}

func (sender *FlatSender) Emoji(emojiID int) *FlatSender {
	t := cqcode.Emoji{
		EmojiID: emojiID,
	}
	sender.cache = append(sender.cache, &t)
	return sender
}

func (sender *Sender) Bface(bfaceID int) (Message, error) {
	t := cqcode.Bface{
		BfaceID: bfaceID,
	}
	sender.cache = append(sender.cache, &t)
	return sender.Send()
}

func (sender *FlatSender) Sface(sfaceID int) *FlatSender {
	t := cqcode.Sface{
		SfaceID: sfaceID,
	}
	sender.cache = append(sender.cache, &t)
	return sender
}

func (sender *Sender) Rps() (Message, error) {
	t := cqcode.Rps{
		Type: 0,
	}
	sender.cache = append(sender.cache, &t)
	return sender.Send()
}

func (sender *Sender) Dice() (Message, error) {
	t := cqcode.Dice{
		Type: 0,
	}
	sender.cache = append(sender.cache, &t)
	return sender.Send()
}

func (sender *Sender) Shake(emojiID int) (Message, error) {
	t := cqcode.Emoji{
		EmojiID: emojiID,
	}
	sender.cache = append(sender.cache, &t)
	return sender.Send()
}

func (sender *Sender) Music(music cqcode.Music) (Message, error) {
	sender.cache = append(sender.cache, &music)
	return sender.Send()
}

func (sender *Sender) Share(share cqcode.Share) (Message, error) {
	sender.cache = append(sender.cache, &share)
	return sender.Send()
}

func (sender *Sender) Location(loc cqcode.Location) (Message, error) {
	sender.cache = append(sender.cache, &loc)
	return sender.Send()
}

func (sender *Sender) Show(id int) (Message, error) {
	t := cqcode.Show{
		ID: id,
	}
	sender.cache = append(sender.cache, &t)
	return sender.Send()
}

func (sender *Sender) Sign(sign cqcode.Sign) (Message, error) {
	sender.cache = append(sender.cache, &sign)
	return sender.Send()
}
