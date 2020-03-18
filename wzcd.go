package wzcd

import (
	"log"
	"time"

	wzlib "github.com/infra-whizz/wzlib"
	wzlib_transport "github.com/infra-whizz/wzlib/transport"
	"github.com/nats-io/nats.go"
)

type WzChannels struct {
	response *nats.Subscription
	console  *nats.Subscription
}

type WzcDaemon struct {
	WzcDaemonEvents
	transport *wzlib_transport.WzdPubSub
	channels  *WzChannels
}

func NewWzcDaemon() *WzcDaemon {
	wz := new(WzcDaemon)
	wz.transport = wzlib_transport.NewWizPubSub()
	wz.channels = new(WzChannels)
	return wz
}

func (wz *WzcDaemon) GetTransport() *wzlib_transport.WzdPubSub {
	return wz.transport
}

// Run the daemon, prior setting it up.
func (wz *WzcDaemon) Run() *WzcDaemon {
	var err error

	wz.GetTransport().Start()
	wz.channels.console, err = wz.GetTransport().GetSubscriber().Subscribe(wzlib.CHANNEL_CONSOLE, wz.onConsoleEvent)
	if err != nil {
		log.Panicf("Unable to subscribe to a console channel: %s\n", err.Error())
	}

	wz.channels.response, err = wz.GetTransport().GetSubscriber().Subscribe(wzlib.CHANNEL_RESPONSE, wz.onResponseEvent)
	if err != nil {
		log.Panicf("Unable to subscribe to a console channel: %s\n", err.Error())
	}

	return wz
}

func (wz *WzcDaemon) AppLoop() {
	for {
		time.Sleep(10 * time.Second)
	}
}
