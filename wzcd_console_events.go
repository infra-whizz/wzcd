package wzcd

import (
	wzlib "github.com/infra-whizz/wzlib"
	wzlib_logger "github.com/infra-whizz/wzlib/logger"
	wzlib_transport "github.com/infra-whizz/wzlib/transport"
)

type WzConsoleEvents struct {
	dispatcher *WzcDaemonDispatcher
	wzlib_logger.WzLogger
}

func NewWzConsoleEvents(dispatcher *WzcDaemonDispatcher) *WzConsoleEvents {
	wce := new(WzConsoleEvents)
	wce.dispatcher = dispatcher
	return wce
}

// Search clients and send back the result
func (wz *WzConsoleEvents) searchClients(query string) {
	wz.GetLogger().Infoln("Sarching for clients")
	wz.GetLogger().Debugf("Serch query: '%s'", query)

	found := wz.dispatcher.daemon.GetDb().GetControllerAPI().GetClientsAPI().Search(query)

	// XXX - refactor - repeating code
	envelope := wzlib_transport.NewWzMessage(wzlib_transport.MSGTYPE_CLIENT)
	envelope.Payload[wzlib_transport.PAYLOAD_BATCH_SIZE] = 1
	envelope.Payload[wzlib_transport.PAYLOAD_FUNC_RET] = map[string]interface{}{"clients.found": found}

	wz.dispatcher.daemon.GetTransport().PublishEnvelopeToChannel(wzlib.CHANNEL_CONTROLLER, envelope)
}

func (wz *WzConsoleEvents) acceptNewClients(fingerprints []interface{}) {
	wz.GetLogger().Infoln("Accepting clients")

	// XXX - refactor - fingerprints: interface to string
	fp := make([]string, len(fingerprints))
	for idx, f := range fingerprints {
		fp[idx] = f.(string)
	}
	missing := wz.dispatcher.daemon.GetDb().GetControllerAPI().GetClientsAPI().Accept(fp...)

	envelope := wzlib_transport.NewWzMessage(wzlib_transport.MSGTYPE_CLIENT)
	envelope.Payload[wzlib_transport.PAYLOAD_BATCH_SIZE] = 1
	envelope.Payload[wzlib_transport.PAYLOAD_FUNC_RET] = map[string]interface{}{"accepted.missing": missing}

	// send
	wz.dispatcher.daemon.GetTransport().PublishEnvelopeToChannel(wzlib.CHANNEL_CONTROLLER, envelope)
}

func (wz *WzConsoleEvents) deleteClients(fingerprints []interface{}) {
	wz.GetLogger().Infoln("Deleting clients")

	// XXX - refactor - fingerprints: interface to string
	fp := make([]string, len(fingerprints))
	for idx, f := range fingerprints {
		fp[idx] = f.(string)
	}
	missing := wz.dispatcher.daemon.GetDb().GetControllerAPI().GetClientsAPI().Delete(fp...)

	// XXX - refactor - repeating code
	envelope := wzlib_transport.NewWzMessage(wzlib_transport.MSGTYPE_CLIENT)
	envelope.Payload[wzlib_transport.PAYLOAD_BATCH_SIZE] = 1
	envelope.Payload[wzlib_transport.PAYLOAD_FUNC_RET] = map[string]interface{}{"deleted.missing": missing}

	// send
	wz.dispatcher.daemon.GetTransport().PublishEnvelopeToChannel(wzlib.CHANNEL_CONTROLLER, envelope)
}

func (wz *WzConsoleEvents) rejectClients(fingerprints []interface{}) {
	wz.GetLogger().Infoln("Rejecting clients")

	// XXX - refactor - fingerprints: interface to string
	fp := make([]string, len(fingerprints))
	for idx, f := range fingerprints {
		fp[idx] = f.(string)
	}
	missing := wz.dispatcher.daemon.GetDb().GetControllerAPI().GetClientsAPI().Reject(fp...)

	// XXX - refactor - repeating code
	envelope := wzlib_transport.NewWzMessage(wzlib_transport.MSGTYPE_CLIENT)
	envelope.Payload[wzlib_transport.PAYLOAD_BATCH_SIZE] = 1
	envelope.Payload[wzlib_transport.PAYLOAD_FUNC_RET] = map[string]interface{}{"rejected.missing": missing}

	// send
	wz.dispatcher.daemon.GetTransport().PublishEnvelopeToChannel(wzlib.CHANNEL_CONTROLLER, envelope)
}
