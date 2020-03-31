package wzcd

import (
	"log"

	wzlib_database_controller "github.com/infra-whizz/wzlib/database/controller"

	"github.com/davecgh/go-spew/spew"
	wzcd_events "github.com/infra-whizz/wzcd/events"
	"github.com/infra-whizz/wzlib"
	wzlib_transport "github.com/infra-whizz/wzlib/transport"
	"github.com/nats-io/nats.go"
)

type WzcDaemonEvents struct {
	daemon *WzcDaemon
}

// NewWzcDaemonEvents creates new instance of the daemon events class
func NewWzcDaemonEvents(daemon *WzcDaemon) *WzcDaemonEvents {
	d := new(WzcDaemonEvents)
	d.daemon = daemon
	return d
}

// OnConsoleEvent receives and dispatches messages on console channel
func (wz *WzcDaemonEvents) OnConsoleEvent(m *nats.Msg) {
	log.Println("On Console channel Event")
	envelope := wzlib_transport.NewWzEventMsgUtils().GetMessage(m.Data)

	switch envelope.Type {
	default:
		log.Println("Discarding unknown message from console channel:")
		spew.Dump(envelope)
	}
}

// OnClientEvent receives and dispatches messages on client channel
func (wz *WzcDaemonEvents) OnClientEvent(m *nats.Msg) {
	log.Println("On Client channel Event")
	envelope := wzlib_transport.NewWzEventMsgUtils().GetMessage(m.Data)

	switch envelope.Type {
	case wzlib_transport.MSGTYPE_PING:
		response := wzlib_transport.NewWzMessage(wzlib_transport.MSGTYPE_PING)
		response.Payload = envelope.Payload
		msg, _ := response.Serialise()
		err := wz.daemon.GetTransport().GetPublisher().Publish(wzlib.CHANNEL_CONTROLLER, msg)
		if err != nil {
			log.Println("Error sending ping event:", err.Error())
		}
	case wzlib_transport.MSGTYPE_REGISTRATION:
		log.Println("Registration event")
	default:
		log.Println("Discarding unknown message from client channel:")
		spew.Dump(envelope)
	}
}
