package natserver_test

import (
	"github.com/nats-io/gnatsd/server"
	natsserver "github.com/nats-io/nats-server/test"
)

func RunServerWithOptions(opts *server.Options) *server.Server {
	return natsserver.RunServer(opts)
}

func RunNATsServerOnPort(port int) *server.Server {
	opts := natsserver.DefaultTestOptions
	opts.Port = port
	return RunServerWithOptions(&opts)
}
