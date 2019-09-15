package qqbotapi

import (
	"fmt"
	"reflect"
)

type Ev struct {
	updatesChannel UpdatesChannel
	subscribers    map[string][]func(update Update)
}

func NewEv(channel UpdatesChannel) *Ev {
	ev := &Ev{
		updatesChannel: channel,
		subscribers:    make(map[string][]func(update Update)),
	}
	go func() {
		for update := range channel {
			postType := update.PostType
			var detailedType string
			switch postType {
			case "notice":
				detailedType = update.NoticeType
			case "message":
				detailedType = update.MessageType
			case "request":
				detailedType = update.RequestType
			}
			if detailedType != "" {
				if update.SubType != "" {
					ev.Emit(
						fmt.Sprintf("%s.%s.%s", postType, detailedType, update.SubType),
						update,
					)
				}
				ev.Emit(
					fmt.Sprintf("%s.%s", postType, detailedType),
					update,
				)
			}
			ev.Emit(postType, update)
		}
	}()
	return ev
}

type Unsubscribe func()

func (ev *Ev) Emit(event string, update Update) {
	if handlers, ok := ev.subscribers[event]; ok {
		for _, handler := range handlers {
			handler(update)
		}
	}
}

func (ev *Ev) On(event string) func(func(update Update)) Unsubscribe {
	return func(handler func(update Update)) Unsubscribe {
		handlers, ok := ev.subscribers[event]
		if !ok {
			ev.subscribers[event] = make([]func(update Update), 0)
			handlers = ev.subscribers[event]
		}
		ev.subscribers[event] = append(handlers, handler)
		return func() {
			ev.Off(event)(handler)
		}
	}
}

func (ev *Ev) Off(event string) func(func(update Update)) {
	return func(handler func(update Update)) {
		handlers, ok := ev.subscribers[event]
		if !ok {
			return
		}
		newHandlers := make([]func(update Update), 0)
		for _, h := range handlers {
			if reflect.ValueOf(h) != reflect.ValueOf(handler) {
				newHandlers = append(newHandlers, h)
			}
		}
		ev.subscribers[event] = newHandlers
	}
}
