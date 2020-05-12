package wzcd

import (
	"time"

	wzlib "github.com/infra-whizz/wzlib"
	wzlib_database "github.com/infra-whizz/wzlib/database"
	wzlib_logger "github.com/infra-whizz/wzlib/logger"
	wzlib_transport "github.com/infra-whizz/wzlib/transport"
	"github.com/nats-io/nats.go"
)

type WzChannels struct {
	clients *nats.Subscription
	console *nats.Subscription
}

type WzcDaemon struct {
	dispatcher *WzcDaemonDispatcher
	transport  *wzlib_transport.WzdPubSub
	channels   *WzChannels
	db         *wzlib_database.WzDBH
	keymanager *WzcPKIManager
	wzlib_logger.WzLogger
}

func NewWzcDaemon() *WzcDaemon {
	wz := new(WzcDaemon)
	wz.transport = wzlib_transport.NewWizPubSub()
	wz.channels = new(WzChannels)
	wz.dispatcher = NewWzcDaemonDispatcher(wz)
	wz.db = wzlib_database.NewWzDBH().WithControllerAPI()
	wz.keymanager = NewWzcPKIManager().SetDbh(wz.db)
	return wz
}

// GetDb connection
func (wz *WzcDaemon) GetDb() *wzlib_database.WzDBH {
	return wz.db
}

// GetTransport for the MQ
func (wz *WzcDaemon) GetTransport() *wzlib_transport.WzdPubSub {
	return wz.transport
}

// GetPkiManager for manage keys locally to the cluster database
func (wz *WzcDaemon) GetPKIManager() *WzcPKIManager {
	return wz.keymanager
}

// Run the daemon, prior setting it up.
func (wz *WzcDaemon) Run() *WzcDaemon {
	var err error

	// Subscribe to the console channel
	wz.GetTransport().Start()
	wz.channels.console, err = wz.GetTransport().
		GetSubscriber().
		Subscribe(wzlib.CHANNEL_CONSOLE, wz.dispatcher.OnConsoleEvent)
	if err != nil {
		wz.GetLogger().Panicf("Unable to subscribe to a console channel: %s\n", err.Error())
	}

	// Subscribe to the response channel
	wz.channels.clients, err = wz.GetTransport().
		GetSubscriber().
		Subscribe(wzlib.CHANNEL_CLIENT, wz.dispatcher.OnClientEvent)
	if err != nil {
		wz.GetLogger().Panicf("Unable to subscribe to a response channel: %s\n", err.Error())
	}

	// Open DB
	wz.GetDb().Open()

	return wz
}

func (wz *WzcDaemon) AppLoop() {
	for {
		time.Sleep(10 * time.Second)
	}
}
