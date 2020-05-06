package wzcd

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/infra-whizz/wzlib"
	wzlib_database_controller "github.com/infra-whizz/wzlib/database/controller"
	wzlib_logger "github.com/infra-whizz/wzlib/logger"
	wzlib_transport "github.com/infra-whizz/wzlib/transport"
	"github.com/nats-io/nats.go"
)

type WzcDaemonDispatcher struct {
	daemon  *WzcDaemon
	console *WzConsoleEvents
	wzlib_logger.WzLogger
}

// NewWzcDaemonDispatcher creates new instance of the daemon events class
func NewWzcDaemonDispatcher(daemon *WzcDaemon) *WzcDaemonDispatcher {
	d := new(WzcDaemonDispatcher)
	d.daemon = daemon
	d.console = NewWzConsoleEvents(d)
	return d
}

// OnConsoleEvent receives and dispatches messages on console channel
func (wz *WzcDaemonDispatcher) OnConsoleEvent(m *nats.Msg) {
	wz.GetLogger().Debugln("On Console channel Event")
	envelope := wzlib_transport.NewWzEventMsgUtils().GetMessage(m.Data)
	spew.Dump(envelope)
	switch envelope.Type {
	case wzlib_transport.MSGTYPE_CLIENT:
		command, ok := envelope.Payload[wzlib_transport.PAYLOAD_COMMAND]
		if !ok {
			wz.GetLogger().Debugln("Discarding console message: unknown command")
			return
		}

		switch command {
		case "list.clients.new":
			go wz.sendListClientsNew()
		case "list.clients.rejected":
			go wz.sendListClientsRejected()
		case "clients.accept":
			params := envelope.Payload[wzlib_transport.PAYLOAD_COMMAND_PARAMS]
			if params != nil {
				fingerprints := params.(map[string]interface{})["fingerprints"]
				if !ok {
					wz.GetLogger().Errorln("Discarding request to accept clients: unspecified target")
				} else {
					if fingerprints != nil {
						go wz.console.acceptNewClients(fingerprints.([]interface{}))
					} else {
						go wz.console.acceptNewClients(make([]interface{}, 0))
					}
				}
			}
		case "clients.reject":
			params := envelope.Payload[wzlib_transport.PAYLOAD_COMMAND_PARAMS]
			if params != nil {
				fingerprints := params.(map[string]interface{})["fingerprints"]
				if fingerprints == nil {
					wz.GetLogger().Errorln("Discarding request to reject clients: unspecified target")
					go wz.console.rejectClients(make([]interface{}, 0))
				} else {
					go wz.console.rejectClients(fingerprints.([]interface{}))
				}
			}
		case "clients.delete":
			params := envelope.Payload[wzlib_transport.PAYLOAD_COMMAND_PARAMS]
			if params != nil {
				fingerprints := params.(map[string]interface{})["fingerprints"]
				if fingerprints == nil {
					wz.GetLogger().Errorln("Discarding request to reject clients: unspecified target")
				} else {
					go wz.console.deleteClients(fingerprints.([]interface{}))
				}
			}
		case "clients.search":
			params := envelope.Payload[wzlib_transport.PAYLOAD_COMMAND_PARAMS]
			if params != nil {
				query := params.(map[string]interface{})["query"]
				if query == nil || query.(string) == "" {
					wz.GetLogger().Errorln("Discarding search request: unspecified query")
				} else {
					go wz.console.searchClients(query.(string))
				}
			}
		default:
			wz.GetLogger().Debugln("Discarding console message: unsupported command -", command)
		}
	default:
		wz.GetLogger().Debugln("Discarding unknown message from console channel:")
	}
}

func (wz *WzcDaemonDispatcher) sendListClientsNew() {
	// call db stuff, obtain everything
	registered := wz.daemon.GetDb().GetControllerAPI().GetClientsAPI().GetRegistered()

	// TODO: Construct batch of messages and send them one by one
	// NATS should run in streaming mode instead (!!)

	// XXX - refactor - repeating code
	envelope := wzlib_transport.NewWzMessage(wzlib_transport.MSGTYPE_CLIENT)
	envelope.Payload[wzlib_transport.PAYLOAD_BATCH_SIZE] = 1
	envelope.Payload[wzlib_transport.PAYLOAD_FUNC_RET] = map[string]interface{}{"registered": registered}

	// send
	wz.daemon.GetTransport().PublishEnvelopeToChannel(wzlib.CHANNEL_CONTROLLER, envelope)
}

func (wz *WzcDaemonDispatcher) sendListClientsRejected() {
	rejected := wz.daemon.GetDb().GetControllerAPI().GetClientsAPI().GetRejected()

	// XXX - refactor - repeating code
	envelope := wzlib_transport.NewWzMessage(wzlib_transport.MSGTYPE_CLIENT)
	envelope.Payload[wzlib_transport.PAYLOAD_BATCH_SIZE] = 1
	envelope.Payload[wzlib_transport.PAYLOAD_FUNC_RET] = map[string]interface{}{"rejected": rejected}

	// send
	wz.daemon.GetTransport().PublishEnvelopeToChannel(wzlib.CHANNEL_CONTROLLER, envelope)
}

// OnClientEvent receives and dispatches messages on client channel
func (wz *WzcDaemonDispatcher) OnClientEvent(m *nats.Msg) {
	wz.GetLogger().Debugln("On Client channel Event")
	envelope := wzlib_transport.NewWzEventMsgUtils().GetMessage(m.Data)

	switch envelope.Type {
	case wzlib_transport.MSGTYPE_PING:
		response := wzlib_transport.NewWzMessage(wzlib_transport.MSGTYPE_PING)
		response.Payload = envelope.Payload
		wz.daemon.GetTransport().PublishEnvelopeToChannel(wzlib.CHANNEL_CONTROLLER, response)
	case wzlib_transport.MSGTYPE_REGISTRATION:
		wz.GetLogger().Debugln("Registering new client")
		wz.registerNewClient(envelope)
	default:
		wz.GetLogger().Debugln("Discarding unknown message from client channel:")
		spew.Dump(envelope)
	}
}

func (wz *WzcDaemonDispatcher) registerNewClient(envelope *wzlib_transport.WzGenericMessage) {
	status := wz.daemon.GetDb().GetControllerAPI().GetClientsAPI().Register(wzlib_database_controller.NewWzClientFromPayload(envelope.Payload))
	response := wzlib_transport.NewWzMessage(wzlib_transport.MSGTYPE_REGISTRATION)
	response.Payload[wzlib_transport.PAYLOAD_FUNC_RET] = map[string]interface{}{"status": status}
	wz.daemon.GetTransport().PublishEnvelopeToChannel(wzlib.CHANNEL_CONTROLLER, response)
}
