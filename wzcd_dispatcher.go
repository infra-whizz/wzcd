package wzcd

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/infra-whizz/wzlib"
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

func (wz *WzcDaemonDispatcher) verifyIncomingMessage(msg *wzlib_transport.WzGenericMessage) (bool, error) {
	sig, ok := msg.Payload[wzlib_transport.PAYLOAD_RSA_SIGNATURE]
	if !ok {
		return false, fmt.Errorf("Message signature not found")
	}
	fp, ok := msg.Payload[wzlib_transport.PAYLOAD_RSA_FINGERPRINT]
	if !ok {
		return false, fmt.Errorf("Message pubkey fingerprint was not found")
	}

	pubkey := wz.daemon.GetDb().GetControllerAPI().GetKeysAPI().GetRSAPublicPEMByFingerprint(fp.(string))
	if pubkey == nil {
		return false, fmt.Errorf("No public key is registered by this fingerprint")
	}

	return wz.daemon.GetCryptoBundle().GetRSA().VerifyPem(pubkey, msg.GetSignableMessageContent(), sig.([]byte))
}

// OnConsoleEvent receives and dispatches messages on console channel
func (wz *WzcDaemonDispatcher) OnConsoleEvent(m *nats.Msg) {
	wz.GetLogger().Debugln("On Console channel Event")
	envelope := wzlib_transport.NewWzEventMsgUtils().GetMessage(m.Data)

	if ret, err := wz.verifyIncomingMessage(envelope); !ret {
		wz.GetLogger().Errorf("Discarding unauthorised message: %s", err.Error())
		wz.console.sendError(fmt.Sprintf("Unauthorised. Access denied: %s", err.Error()))
		return
	}

	spew.Dump(envelope)
	switch envelope.Type {
	case wzlib_transport.MSGTYPE_CLIENT:
		command, ok := envelope.Payload[wzlib_transport.PAYLOAD_COMMAND]
		if !ok {
			wz.GetLogger().Debugln("Discarding console message: unknown command")
			wz.console.sendError("Unknown command.")
			return
		}

		switch command {
		case "list.clients.new":
			go wz.console.sendListClientsNew()
		case "list.clients.rejected":
			go wz.console.sendListClientsRejected()
		case "clients.accept":
			wz.action_clientsAccept(envelope)
		case "clients.reject":
			wz.action_clientsReject(envelope)
		case "clients.delete":
			wz.action_clientsDelete(envelope)
		case "clients.search":
			wz.action_clientsSearch(envelope)
		default:
			wz.GetLogger().Debugln("Discarding console message: unsupported command -", command)
		}
	default:
		wz.GetLogger().Debugln("Discarding unknown message from console channel:")
	}
}

// Perform action of "client.accept"
func (wz *WzcDaemonDispatcher) action_clientsAccept(envelope *wzlib_transport.WzGenericMessage) {
	params := envelope.Payload[wzlib_transport.PAYLOAD_COMMAND_PARAMS]
	if params != nil {
		fingerprints := params.(map[string]interface{})["fingerprints"]
		if fingerprints != nil {
			go wz.console.acceptNewClients(fingerprints.([]interface{}))
		} else {
			go wz.console.acceptNewClients(make([]interface{}, 0))
		}
	} else {
		wz.GetLogger().Errorln("No params has been specified when accepting clients")
	}
}

// Perform action of "client.reject"
func (wz *WzcDaemonDispatcher) action_clientsReject(envelope *wzlib_transport.WzGenericMessage) {
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
}

// Perform action of "client.delete"
func (wz *WzcDaemonDispatcher) action_clientsDelete(envelope *wzlib_transport.WzGenericMessage) {
	params := envelope.Payload[wzlib_transport.PAYLOAD_COMMAND_PARAMS]
	if params != nil {
		fingerprints := params.(map[string]interface{})["fingerprints"]
		if fingerprints == nil {
			wz.GetLogger().Errorln("Discarding request to reject clients: unspecified target")
		} else {
			go wz.console.deleteClients(fingerprints.([]interface{}))
		}
	}
}

// Perform action of "client.search"
func (wz *WzcDaemonDispatcher) action_clientsSearch(envelope *wzlib_transport.WzGenericMessage) {
	params := envelope.Payload[wzlib_transport.PAYLOAD_COMMAND_PARAMS]
	if params != nil {
		query := params.(map[string]interface{})["query"]
		if query == nil || query.(string) == "" {
			wz.GetLogger().Errorln("Discarding search request: unspecified query")
		} else {
			go wz.console.searchClients(query.(string))
		}
	}
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
		wz.console.registerNewClient(envelope)
	default:
		wz.GetLogger().Debugln("Discarding unknown message from client channel:")
		spew.Dump(envelope)
	}
}
