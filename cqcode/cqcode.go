// Package cqcode provides basic structs of cqcode media, and utilities of parsing
// and formatting cqcode
package cqcode

import (
	"strings"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"regexp"
	"fmt"
	"reflect"
	"strconv"
)

// StrictCommand indicates that whether a command must start with "/".
// See function #Command
var StrictCommand = false

// A Message is a sort of Media.
type Message []Media

// A MessageSegment is a struct which has "type" and "data", see documentation at
// https://cqhttp.cc/docs/3.4/#/Message?id=%E6%B6%88%E6%81%AF%E6%AE%B5%EF%BC%88%E5%B9%BF%E4%B9%89-cq-%E7%A0%81%EF%BC%89
type MessageSegment struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

func (seg *MessageSegment) FunctionName() string {
	return seg.Type
}

// NewMessageSegment formats MessageSegment from any type of Media.
func NewMessageSegment(media Media) (MessageSegment, error) {
	seg := MessageSegment{}
	seg.Type = media.FunctionName()
	seg.Data = make(map[string]interface{})
	err := decode(media, &seg.Data)
	return seg, err
}

// NewMessageSegmentFromCQCode parses a CQCode to a NewMessageSegment.
func NewMessageSegmentFromCQCode(str string) (MessageSegment, error) {
	seg := MessageSegment{}
	seg.Data = make(map[string]interface{})
	err := ParseCQCode(str, &seg)
	return seg, err
}

func decode(input, output interface{}) error {
	config := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           output,
		WeaklyTypedInput: true,
		TagName:          "cq",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

// NewMessage returns an empty Message.
func NewMessage() (Message) {
	return make(Message, 0)
}

// ParseMessageSegments parses msg, which might have 2 types, string or array,
// depending on the configuration of cqhttp, to a sort of MessageSegment.
// msg is the value of key "message" of the data umarshalled from the
// API response JSON.
func ParseMessageSegments(msg interface{}) ([]MessageSegment, error) {
	switch x := msg.(type) {
	case string:
		return ParseMessageSegmentsFromString(x)
	default:
		return ParseMessageSegmentsFromArray(x)
	}
}

// ParseMessage parses msg, which might have 2 types, string or array,
// depending on the configuration of cqhttp, to a Message.
// msg is the value of key "message" of the data umarshalled from the
// API response JSON.
func ParseMessage(msg interface{}) (Message, error) {
	switch x := msg.(type) {
	case string:
		return ParseMessageFromString(x)
	default:
		return ParseMessageFromArray(x)
	}
}

// ParseMessageSegmentsFromArray parses msg as type array to a sort of MessageSegment.
// msg is the value of key "message" of the data umarshalled from the
// API response JSON.
func ParseMessageSegmentsFromArray(msg interface{}) ([]MessageSegment, error) {
	segs := make([]MessageSegment, 0)
	err := decode(msg, segs)
	return segs, err
}

// ParseMessageFromArray parses msg as type array to a Message.
// msg is the value of key "message" of the data umarshalled from the
// API response JSON.
func ParseMessageFromArray(msg interface{}) (Message, error) {
	segs, err := ParseMessageSegmentsFromArray(msg)
	if err != nil {
		return NewMessage(), err
	}
	return ParseMessageFromMessageSegments(segs), nil
}

// ParseMessageSegmentsFromString parses msg as type string to a sort of MessageSegment.
// msg is the value of key "message" of the data umarshalled from the
// API response JSON.
func ParseMessageSegmentsFromString(str string) ([]MessageSegment, error) {
	segs := make([]MessageSegment, 0)
	res := regexp.MustCompile(`\[CQ:[\s\S]*?\]`).FindAllStringSubmatchIndex(str, -1)
	i := 0
	for _, cqc := range res {
		if cqc[0] > i {
			// There is a text message before this cqc
			seg := MessageSegment{
				Type: "text",
				Data: map[string]interface{}{
					"text": DecodeCQCodeText(str[i:cqc[0]]),
				},
			}
			segs = append(segs, seg)
		}
		i = cqc[1]
		seg, err := NewMessageSegmentFromCQCode(str[cqc[0]:cqc[1]])
		if err != nil {
			continue
		}
		segs = append(segs, seg)
	}
	if len(str) > i {
		// There is a text message after all cqc
		seg := MessageSegment{
			Type: "text",
			Data: map[string]interface{}{
				"text": DecodeCQCodeText(str[i:]),
			},
		}
		segs = append(segs, seg)
	}
	return segs, nil
}

// ParseMessageFromString parses msg as type string to a Message.
// msg is the value of key "message" of the data umarshalled from the
// API response JSON.
func ParseMessageFromString(str string) (Message, error) {
	segs, _ := ParseMessageSegmentsFromString(str)
	return ParseMessageFromMessageSegments(segs), nil
}

// ParseMessageFromMessageSegments parses a sort of MessageSegment to a Message.
func ParseMessageFromMessageSegments(segs []MessageSegment) Message {
	message := NewMessage()
	for _, seg := range segs {
		switch seg.Type {
		case "text":
			text := Text{}
			seg.ParseMedia(&text)
			message = append(message, &text)
		case "at":
			at := At{}
			seg.ParseMedia(&at)
			message = append(message, &at)
		case "face":
			face := Face{}
			seg.ParseMedia(&face)
			message = append(message, &face)
		case "emoji":
			emoji := Emoji{}
			seg.ParseMedia(&emoji)
			message = append(message, &emoji)
		case "bface":
			bface := Bface{}
			seg.ParseMedia(&bface)
			message = append(message, &bface)
		case "sface":
			sface := Sface{}
			seg.ParseMedia(&sface)
			message = append(message, &sface)
		case "image":
			image := Image{}
			seg.ParseMedia(&image)
			message = append(message, &image)
		case "record":
			record := Record{}
			seg.ParseMedia(&record)
			message = append(message, &record)
		case "rps":
			rps := Rps{}
			seg.ParseMedia(&rps)
			message = append(message, &rps)
		case "dice":
			dice := Dice{}
			seg.ParseMedia(&dice)
			message = append(message, &dice)
		case "shake":
			shake := Shake{}
			seg.ParseMedia(&shake)
			message = append(message, &shake)
		case "music":
			music := Music{}
			seg.ParseMedia(&music)
			message = append(message, &music)
		case "share":
			share := Share{}
			seg.ParseMedia(&share)
			message = append(message, &share)
		case "location":
			location := Location{}
			seg.ParseMedia(&location)
			message = append(message, &location)
		case "show":
			show := Show{}
			seg.ParseMedia(&show)
			message = append(message, &show)
		case "sign":
			sign := Sign{}
			seg.ParseMedia(&sign)
			message = append(message, &sign)
		case "rich":
			rich := Rich{}
			seg.ParseMedia(&rich)
			message = append(message, &rich)
		default:
			s := seg
			message = append(message, &s)
		}
	}
	return message
}

// IsCommand indicates whether a Message is a command.
// If StrictCommand is true, only messages start with "/" will be regard as command.
func (m *Message) IsCommand() bool {
	str := m.CQString()
	return IsCommand(str)
}

// Command parses a command message and returns the command with command arguments.
// In a StrictCommand mode, the initial "/" in a command will be stripped off.
func (m *Message) Command() (cmd string, args []string) {
	str := m.CQString()
	return Command(str)
}

// IsCommand indicates whether a string is a command.
// If StrictCommand is true, only strings start with "/" will be regard as command.
func IsCommand(str string) bool {
	if len(str) == 0 {
		return false
	}
	if StrictCommand && str[:1] != "/" {
		return false
	}
	return true
}

// Command parses a command string and returns the command with command arguments.
// In a StrictCommand mode, the initial "/" in a command will be stripped off.
func Command(str string) (cmd string, args []string) {
	str = strings.Replace(str, `\\`, `\0x5c`, -1)
	str = strings.Replace(str, `\"`, `\0x22`, -1)
	str = strings.Replace(str, `\'`, `\0x27`, -1)
	strs := regexp.MustCompile(`'[\s\S]*?'|"[\s\S]*?"|\S*\[CQ:[\s\S]*?\]\S*|\S+`).FindAllString(str, -1)
	if len(strs) == 0 || len(strs[0]) == 0 {
		return
	}
	if StrictCommand {
		if strs[0][:1] != "/" {
			return
		}
		cmd = strs[0][1:]
	} else {
		cmd = strs[0]
	}
	for _, arg := range strs[1:] {
		arg = strings.Trim(arg, `'"`)
		arg = strings.Replace(arg, `\0x27`, `'`, -1)
		arg = strings.Replace(arg, `\0x22`, `"`, -1)
		arg = strings.Replace(arg, `\0x5c`, `\`, -1)
		args = append(args, arg)
	}
	return
}

// CQString returns the CQEncoded string. All media in the message will be converted
// to its CQCode.
func (m *Message) CQString() string {
	var str string
	for _, media := range *m {
		str += FormatCQCode(media)
	}
	return str
}

// MessageSegments returns an array of MessageSegment, you will find this useful if you
// configured your cqhttp to receive messages in type of array.
func (m *Message) MessageSegments() []MessageSegment {
	segs := make([]MessageSegment, 0)
	for _, media := range *m {
		seg, err := NewMessageSegment(media)
		if err != nil {
			continue
		}
		segs = append(segs, seg)
	}
	return segs
}

// Append is just an alias to append, which appends media to m.
func (m *Message) Append(media Media) error {
	*m = append(*m, media)
	return nil
}

// ParseMedia parses a MessageSegment to a specified type of Media.
func (seg *MessageSegment) ParseMedia(media Media) error {
	_, ok := media.(*MessageSegment)
	if ok {
		reflect.ValueOf(media).Elem().Set(reflect.ValueOf(seg).Elem())
		return nil
	}
	err := decode(seg.Data, media)
	if seg.Type != media.FunctionName() {
		if err != nil {
			err = errors.Wrap(err, "wrong media type")
		} else {
			err = errors.New("wrong media type")
		}
	}
	return err
}

// ParseMedia parses a CQEncoded string to a specified type of Media.
func ParseCQCode(str string, media Media) (error) {
	l := len(str)
	if l <= 5 || str[:4] != "[CQ:" || str[len(str)-1:] != "]" {
		// Invalid CQCode
		switch v := media.(type) {
		case *Text:
			v.Text = DecodeCQText(str)
			return nil
		case *MessageSegment:
			v.Type = "text"
			v.Data = map[string]interface{}{
				"text": DecodeCQText(str),
			}
			return nil
		default:
			err := errors.New("invalid cqcode")
			return err
		}
	}
	str = str[4 : len(str)-1]
	strs := strings.Split(str, ",")
	ms := MessageSegment{
		Type: strs[0],
		Data: make(map[string]interface{}),
	}
	for i := 1; i < len(strs); i++ {
		kvstrs := strings.Split(strs[i], "=")
		if len(kvstrs) == 0 {
			continue
		}
		ms.Data[kvstrs[0]] = DecodeCQCodeText(strings.Join(kvstrs[1:], "="))
	}
	err := ms.ParseMedia(media)
	return err
}

// CQString returns the CQCode of a MessageSegment.
func (seg *MessageSegment) CQString() string {
	return FormatCQCode(seg)
}

// FormatCQCode returns the CQCode of a Media.
func FormatCQCode(media Media) string {
	switch v := media.(type) {
	case *MessageSegment:
		if v.Type == "text" {
			t, ok := v.Data["text"]
			if !ok {
				return ""
			}
			text := fmt.Sprint(t)
			text = EncodeCQText(text)
			return text
		}
		strs := make([]string, 0)
		strs = append(strs, v.Type)
		for k, v := range v.Data {
			text := fmt.Sprint(v)
			text = EncodeCQCodeText(text)
			kvs := fmt.Sprintf("%s=%s", k, text)
			strs = append(strs, kvs)
		}
		str := strings.Join(strs, ",")
		str = fmt.Sprintf("[CQ:%s]", str)
		return str
	case *Text:
		text := EncodeCQText(v.Text)
		return text
	default:
		rv := reflect.ValueOf(v)
		rv = reflect.Indirect(rv)
		if rv.Kind() != reflect.Struct {
			return ""
		}
		strs := make([]string, 0)
		strs = append(strs, v.FunctionName())
		for i := 0; i < rv.NumField(); i++ {
			f := rv.Type().Field(i)
			k := f.Tag.Get("cq")
			if k == "" {
				k = f.Name
			}
			text := fmt.Sprint(rv.Field(i))
			text = EncodeCQCodeText(text)
			kvs := fmt.Sprintf("%s=%s", k, text)
			strs = append(strs, kvs)
		}
		str := strings.Join(strs, ",")
		str = fmt.Sprintf("[CQ:%s]", str)
		return str
	}
}

// Media is any kind of media that could be contained in a message.
type Media interface {
	// FunctionName returns the "function name" defined by Coolq, see documentation at
	// https://d.cqp.me/Pro/CQ%E7%A0%81
	FunctionName() string
}

// Plain text
type Text struct {
	Text string `cq:"text"`
}

func (t *Text) FunctionName() string {
	return "text"
}

// Mention @
type At struct {
	QQ string `cq:"qq"` // Someone's QQ号, could be "all"
}

func (a *At) FunctionName() string {
	return "at"
}

// QQ表情
type Face struct {
	FaceID int `cq:"id"` // 1-170 (旧版), >170 (新表情)
}

func (f *Face) FunctionName() string {
	return "face"
}

// Emoji
type Emoji struct {
	EmojiID int `cq:"id"` // Unicode Dec
}

func (e *Emoji) FunctionName() string {
	return "emoji"
}

// 原创表情
type Bface struct {
	BfaceID int `cq:"id"`
}

func (b *Bface) FunctionName() string {
	return "bface"
}

// 小表情
type Sface struct {
	SfaceID int `cq:"id"`
}

func (s *Sface) FunctionName() string {
	return "sface"
}

// Image
type Image struct {
	FileID string `cq:"file"`
	URL    string `cq:"url"`
}

func (i *Image) FunctionName() string {
	return "image"
}

// Record
type Record struct {
	FileID string `cq:"file"`
	Magic  bool   `cq:"magic"`
	URL    string `cq:"url"`
}

func (r *Record) FunctionName() string {
	return "record"
}

// 猜拳魔法表情
type Rps struct {
	Type int `cq:"type"`
}

// Common value of Rps
const (
	Rock     = 1
	Paper    = 2
	Scissors = 3
)

func (rps *Rps) FunctionName() string {
	return "rps"
}

// 掷骰子魔法表情
type Dice struct {
	Type int `cq:"type"` // 1-6
}

func (d *Dice) FunctionName() string {
	return "dice"
}

// 戳一戳
type Shake struct {
}

func (s *Shake) FunctionName() string {
	return "shake"
}

// 音乐
type Music struct {
	Type string `cq:"type"` // qq, 163, xiami
	// non-custom music
	MusicID string `cq:"id"` // id
	// custom music
	ShareURL string `cq:"url"`     // Link open on click
	AudioURL string `cq:"audio"`   // Link of audio
	Title    string `cq:"title"`   // Title
	Content  string `cq:"content"` // Description
	Image    string `cq:"image"`   // Link of cover image
}

func (m *Music) FunctionName() string {
	return "music"
}

// IsCustomMusic shows whether a music is a custom music.
func (m *Music) IsCustomMusic() bool {
	return m.Type == "custom"
}

// 分享链接
type Share struct {
	URL     string `cq:"url"`
	Title   string `cq:"title"`   // In 12 words
	Content string `cq:"content"` // In 30 words
	Image   string `cq:"image"`   // Link of cover image
}

func (s *Share) FunctionName() string {
	return "share"
}

// 位置
type Location struct {
}

func (l *Location) FunctionName() string {
	return "location"
}

// 厘米秀
type Show struct {
}

func (s *Show) FunctionName() string {
	return "show"
}

// 签到
type Sign struct {
}

func (s *Sign) FunctionName() string {
	return "sign"
}

// 其他富媒体
type Rich struct {
}

func (r *Rich) FunctionName() string {
	return "rich"
}

// EncodeCQText escapes special characters in a non-media plain message.
func EncodeCQText(str string) string {
	str = strings.Replace(str, "&", "&amp;", -1)
	str = strings.Replace(str, "[", "&#91;", -1)
	str = strings.Replace(str, "]", "&#93;", -1)
	return str
}

// DecodeCQText unescapes special characters in a non-media plain message.
func DecodeCQText(str string) string {
	str = strings.Replace(str, "&#93;", "]", -1)
	str = strings.Replace(str, "&#91;", "[", -1)
	str = strings.Replace(str, "&amp;", "&", -1)
	return str
}

// EncodeCQCodeText escapes special characters in a cqcode value.
func EncodeCQCodeText(str string) string {
	str = strings.Replace(str, "&", "&amp;", -1)
	str = strings.Replace(str, "[", "&#91;", -1)
	str = strings.Replace(str, "]", "&#93;", -1)
	str = strings.Replace(str, ",", "&#44;", -1)
	return str
}

// DecodeCQCodeText unescapes special characters in a cqcode value.
func DecodeCQCodeText(str string) string {
	str = strings.Replace(str, "&#44;", ",", -1)
	str = strings.Replace(str, "&#93;", "]", -1)
	str = strings.Replace(str, "&#91;", "[", -1)
	str = strings.Replace(str, "&amp;", "&", -1)
	return str
}

// NewFaceFromName returns a face that corresponds to a given face name.
func NewFaceFromName(str string) (*Face, error) {
	str = strings.Trim(str, "/")
	face := Face{}
	fi, ok := stringFace[str]
	if ok {
		face.FaceID = fi
		return &face, nil
	}
	return &face, errors.New("Unknown face")
}

// Name returns the name of a face
func (f *Face) Name() (string, error) {
	str, ok := faceString[f.FaceID]
	if ok {
		return str, nil
	}
	return strconv.Itoa(f.FaceID), errors.New("Unknown face")
}

var stringFace = map[string]int{
	"微笑":   14,
	"撇嘴":   1,
	"色":    2,
	"发呆":   3,
	"得意":   4,
	"流泪":   5,
	"害羞":   6,
	"闭嘴":   7,
	"睡":    8,
	"大哭":   9,
	"尴尬":   10,
	"发怒":   11,
	"调皮":   12,
	"呲牙":   13,
	"惊讶":   0,
	"难过":   15,
	"酷":    16,
	"冷汗":   96,
	"抓狂":   18,
	"吐":    19,
	"偷笑":   20,
	"可爱":   21,
	"白眼":   22,
	"傲慢":   23,
	"饥饿":   24,
	"困":    25,
	"惊恐":   26,
	"流汗":   27,
	"憨笑":   28,
	"大兵":   29,
	"奋斗":   30,
	"咒骂":   31,
	"疑问":   32,
	"嘘":    33,
	"晕":    34,
	"折磨":   35,
	"衰":    36,
	"骷髅":   37,
	"敲打":   38,
	"再见":   39,
	"擦汗":   97,
	"抠鼻":   98,
	"鼓掌":   99,
	"糗大了":  100,
	"坏笑":   101,
	"左哼哼":  102,
	"右哼哼":  103,
	"哈欠":   104,
	"鄙视":   105,
	"委屈":   106,
	"快哭了":  107,
	"阴险":   108,
	"亲亲":   109,
	"吓":    110,
	"可怜":   111,
	"眨眼睛":  172,
	"笑哭":   182,
	"doge": 179,
	"泪奔":   173,
	"无奈":   174,
	"托腮":   212,
	"卖萌":   175,
	"斜眼笑":  178,
	"喷血":   177,
	"惊喜":   180,
	"骚扰":   181,
	"小纠结":  176,
	"我最美":  183,
	"菜刀":   112,
	"西瓜":   89,
	"啤酒":   113,
	"篮球":   114,
	"乒乓":   115,
	"茶":    171,
	"咖啡":   60,
	"饭":    61,
	"猪头":   46,
	"玫瑰":   63,
	"凋谢":   64,
	"示爱":   116,
	"爱心":   66,
	"心碎":   67,
	"蛋糕":   53,
	"闪电":   54,
	"炸弹":   55,
	"刀":    56,
	"足球":   57,
	"瓢虫":   117,
	"便便":   59,
	"月亮":   75,
	"太阳":   74,
	"礼物":   69,
	"拥抱":   49,
	"强":    76,
	"弱":    77,
	"握手":   78,
	"胜利":   79,
	"抱拳":   118,
	"勾引":   119,
	"拳头":   120,
	"差劲":   121,
	"爱你":   122,
	"NO":   123,
	"OK":   124,
	"爱情":   42,
	"飞吻":   85,
	"跳跳":   43,
	"发抖":   41,
	"怄火":   86,
	"转圈":   125,
	"磕头":   126,
	"回头":   127,
	"跳绳":   128,
	"挥手":   129,
	"激动":   130,
	"街舞":   131,
	"献吻":   132,
	"左太极":  133,
	"右太极":  134,
	"双喜":   136,
	"鞭炮":   137,
	"灯笼":   138,
	"K歌":   140,
	"喝彩":   144,
	"祈祷":   145,
	"爆筋":   146,
	"棒棒糖":  147,
	"喝奶":   148,
	"飞机":   151,
	"钞票":   158,
	"药":    168,
	"手枪":   169,
	"蛋":    188,
	"红包":   192,
	"河蟹":   184,
	"羊驼":   185,
	"菊花":   190,
	"幽灵":   187,
	"大笑":   193,
	"不开心":  194,
	"冷漠":   197,
	"呃":    198,
	"好棒":   199,
	"拜托":   200,
	"点赞":   201,
	"无聊":   202,
	"托脸":   203,
	"吃":    204,
	"送花":   205,
	"害怕":   206,
	"花痴":   207,
	"小样儿":  208,
	"飙泪":   210,
	"我不看":  211,
}

var faceString = map[int]string{
	14:  "微笑",
	1:   "撇嘴",
	2:   "色",
	3:   "发呆",
	4:   "得意",
	5:   "流泪",
	6:   "害羞",
	7:   "闭嘴",
	8:   "睡",
	9:   "大哭",
	10:  "尴尬",
	11:  "发怒",
	12:  "调皮",
	13:  "呲牙",
	0:   "惊讶",
	15:  "难过",
	16:  "酷",
	96:  "冷汗",
	18:  "抓狂",
	19:  "吐",
	20:  "偷笑",
	21:  "可爱",
	22:  "白眼",
	23:  "傲慢",
	24:  "饥饿",
	25:  "困",
	26:  "惊恐",
	27:  "流汗",
	28:  "憨笑",
	29:  "大兵",
	30:  "奋斗",
	31:  "咒骂",
	32:  "疑问",
	33:  "嘘",
	34:  "晕",
	35:  "折磨",
	36:  "衰",
	37:  "骷髅",
	38:  "敲打",
	39:  "再见",
	97:  "擦汗",
	98:  "抠鼻",
	99:  "鼓掌",
	100: "糗大了",
	101: "坏笑",
	102: "左哼哼",
	103: "右哼哼",
	104: "哈欠",
	105: "鄙视",
	106: "委屈",
	107: "快哭了",
	108: "阴险",
	109: "亲亲",
	110: "吓",
	111: "可怜",
	172: "眨眼睛",
	182: "笑哭",
	179: "doge",
	173: "泪奔",
	174: "无奈",
	212: "托腮",
	175: "卖萌",
	178: "斜眼笑",
	177: "喷血",
	180: "惊喜",
	181: "骚扰",
	176: "小纠结",
	183: "我最美",
	112: "菜刀",
	89:  "西瓜",
	113: "啤酒",
	114: "篮球",
	115: "乒乓",
	171: "茶",
	60:  "咖啡",
	61:  "饭",
	46:  "猪头",
	63:  "玫瑰",
	64:  "凋谢",
	116: "示爱",
	66:  "爱心",
	67:  "心碎",
	53:  "蛋糕",
	54:  "闪电",
	55:  "炸弹",
	56:  "刀",
	57:  "足球",
	117: "瓢虫",
	59:  "便便",
	75:  "月亮",
	74:  "太阳",
	69:  "礼物",
	49:  "拥抱",
	76:  "强",
	77:  "弱",
	78:  "握手",
	79:  "胜利",
	118: "抱拳",
	119: "勾引",
	120: "拳头",
	121: "差劲",
	122: "爱你",
	123: "NO",
	124: "OK",
	42:  "爱情",
	85:  "飞吻",
	43:  "跳跳",
	41:  "发抖",
	86:  "怄火",
	125: "转圈",
	126: "磕头",
	127: "回头",
	128: "跳绳",
	129: "挥手",
	130: "激动",
	131: "街舞",
	132: "献吻",
	133: "左太极",
	134: "右太极",
	136: "双喜",
	137: "鞭炮",
	138: "灯笼",
	140: "K歌",
	144: "喝彩",
	145: "祈祷",
	146: "爆筋",
	147: "棒棒糖",
	148: "喝奶",
	151: "飞机",
	158: "钞票",
	168: "药",
	169: "手枪",
	188: "蛋",
	192: "红包",
	184: "河蟹",
	185: "羊驼",
	190: "菊花",
	187: "幽灵",
	193: "大笑",
	194: "不开心",
	197: "冷漠",
	198: "呃",
	199: "好棒",
	200: "拜托",
	201: "点赞",
	202: "无聊",
	203: "托脸",
	204: "吃",
	205: "送花",
	206: "害怕",
	207: "花痴",
	208: "小样儿",
	210: "飙泪",
	211: "我不看",
}
