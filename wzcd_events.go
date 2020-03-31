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
	daemon  *WzcDaemon
	console *wzcd_events.WzConsoleEvents
}

// NewWzcDaemonEvents creates new instance of the daemon events class
func NewWzcDaemonEvents(daemon *WzcDaemon) *WzcDaemonEvents {
	d := new(WzcDaemonEvents)
	d.daemon = daemon
	d.console = wzcd_events.NewWzConsoleEvents()
	return d
}

// OnConsoleEvent receives and dispatches messages on console channel
func (wz *WzcDaemonEvents) OnConsoleEvent(m *nats.Msg) {
	log.Println("On Console channel Event")
	envelope := wzlib_transport.NewWzEventMsgUtils().GetMessage(m.Data)
	switch envelope.Type {
	case wzlib_transport.MSGTYPE_CLIENT:
		spew.Dump(envelope.Payload)
		command, ok := envelope.Payload["command"]
		if !ok {
			log.Println("Discarding console message: unknown command")
			return
		}

		switch command {
		case "list.clients.new":
			go wz.sendListClientsNew()
		case "list.clients.rejected":
			go wz.sendListClientsRejected()
		default:
			log.Println("Discarding console message: unsupported command -", command)
		}
	default:
		log.Println("Discarding unknown message from console channel:")
		spew.Dump(envelope)
	}
}

func (wz *WzcDaemonEvents) sendListClientsNew() {
	log.Println("Get list of the new clients")
	log.Println("Send message[s] back")

	// call db stuff, obtain everything
	//wz.daemon.GetDb().GetControllerAPI().Clients.GetNew()

	// Construct batch of messages and send them one by one
	// NATS should run in streaming mode instead (!!)
	envelope := wzlib_transport.NewWzMessage(wzlib_transport.MSGTYPE_CLIENT)
	envelope.Payload["batch.max"] = 1
	envelope.Payload["clients.new"] = "list of structures here in a future"
	wz.daemon.GetTransport().PublishEnvelopeToChannel(wzlib.CHANNEL_CONTROLLER, envelope)
}

func (wz *WzcDaemonEvents) sendListClientsRejected() {
	log.Println("Get list of the rejected clients")
	log.Println("Send message[s] back")
}

// OnClientEvent receives and dispatches messages on client channel
func (wz *WzcDaemonEvents) OnClientEvent(m *nats.Msg) {
	log.Println("On Client channel Event")
	envelope := wzlib_transport.NewWzEventMsgUtils().GetMessage(m.Data)

	switch envelope.Type {
	case wzlib_transport.MSGTYPE_PING:
		response := wzlib_transport.NewWzMessage(wzlib_transport.MSGTYPE_PING)
		response.Payload = envelope.Payload
		wz.daemon.GetTransport().PublishEnvelopeToChannel(wzlib.CHANNEL_CONTROLLER, response)
	case wzlib_transport.MSGTYPE_REGISTRATION:
		log.Println("Registering new client")
		wz.registerNewClient(envelope)
	default:
		log.Println("Discarding unknown message from client channel:")
		spew.Dump(envelope)
	}
}

func (wz *WzcDaemonEvents) registerNewClient(envelope *wzlib_transport.WzGenericMessage) {
	wz.daemon.GetDb().GetControllerAPI().GetClientsAPI().Register(wzlib_database_controller.NewWzClientFromPayload(envelope.Payload))

	response := wzlib_transport.NewWzMessage(wzlib_transport.MSGTYPE_REGISTRATION)
	response.Payload["registration.status"] = "pending"
	wz.daemon.GetTransport().PublishEnvelopeToChannel(wzlib.CHANNEL_CONTROLLER, response)

}
