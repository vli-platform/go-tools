package notify

import (
	"fmt"
	"log"
)

var notifyEngine NotifyEngine

func init() {
	notifyEngine.Init()
}

type NotifyChannel interface {
	PushMessage(string, ...interface{}) error
}

type Config struct {
	NotifyChannel NotifyChannel
	Prefix        []string
}

type NotifyEngine struct {
	NotifyChannel
	prefix  string
	buffQue chan string
}

func (n *NotifyEngine) Init() {
	n.buffQue = make(chan string, 500)
	go func() {

		for msg := range n.buffQue {
			if n.NotifyChannel == nil {
				continue
			}
			if err := n.NotifyChannel.PushMessage("", msg); err != nil {
				log.Print(err)
			}
		}

	}()
}

func (n *NotifyEngine) PushMessage(format string, a ...interface{}) {
	var msg string
	if format != "" {
		msg = n.prefix + fmt.Sprintf(format, a...)
	} else {
		msg = n.prefix + fmt.Sprintln(a...)
	}

	select {
	case n.buffQue <- msg:
	default: // trường hợp lỗi gửi tin nhắn dẫn đến channel full thì bỏ qua tìn nhắn để tránh lỗi block hết go routines dẫn đến application không thể nhận request tiếp
	}
}

func SetConfig(cfg Config) {
	for _, str := range cfg.Prefix {
		notifyEngine.prefix += str + " / "
	}

	notifyEngine.NotifyChannel = cfg.NotifyChannel
}

func PushMessage(format string, a ...interface{}) {
	notifyEngine.PushMessage(format, a...)
}
