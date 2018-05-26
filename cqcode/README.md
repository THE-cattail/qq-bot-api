## CQ码

[![GoDoc](https://godoc.org/github.com/juzi5201314/cqhttp-go-sdk/cqcode?status.svg)](https://godoc.org/github.com/juzi5201314/cqhttp-go-sdk/cqcode)

有关 CQ 码请参考 [酷Q官方CQ码说明](https://d.cqp.me/Pro/CQ码) 以及 [cq-http插件说明](https://cqhttp.cc/docs/3.4/#/CQCode)

解码消息

```go
func pm(sub_type string, message_id float64, user_id float64, message string, font float64) map[string]interface{} {

	message, err := cqcode.ParseMessage(message)
	if err != nil {
		return map[string]interface{}{}
	}

	for _, m := range message {

		switch x := m.(type) {
		case *cqcode.Image:
			fmt.Print(x.FileID)
		}

	}
	...
}
```

编码消息

```go
...
	m := cqcode.NewMessage()

	face := cqcode.Face{
		FaceID: 170,
	}
	m.Append(&face)

	// 如果消息上报格式为 string 则转换为 string
	messageStr := m.CQString()
	// 如果为 array 则转换为 []MessageSegment
	messageSegments := m.MessageSegments()
...
```

命令解析

```go
func pm(sub_type string, message_id float64, user_id float64, message string, font float64) map[string]interface{} {

	// 命令必须以 "/" 开头，并且解析时自动去掉 "/"，默认为 false
	cqcode.StrictCommand = true

	// 如果上报格式为 string 可以使用静态方法
	if !cqcode.IsCommand(m.(string)) {
		return map[string]interface{}{}
	}
	cmd, args := cqcode.Command(m.(string))

	// 或者先解码为 Message
	m, err := cqcode.ParseMessage(message)
	if err != nil {
		return map[string]interface{}{}
	}
	if !m.IsCommand() {
		return map[string]interface{}{}
	}
	cmd, args := m.Command()

	// cmd string, args []string
	// 注意：cmd 和 args 仍然为富媒体，可以使用 ParseMessage 解析
	...
}
```
