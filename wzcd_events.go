package wzcd

import (
	"log"

	"github.com/nats-io/nats.go"
)

type WzcDaemonEvents struct {
}

func (wz *WzcDaemonEvents) onConsoleEvent(m *nats.Msg) {
	log.Println("received from console", len(m.Data), "bytes")
}

func (wz *WzcDaemonEvents) onResponseEvent(m *nats.Msg) {
	log.Println("received from response channel", len(m.Data), "bytes")
}
