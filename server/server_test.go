package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cuongtranba/faasv"
	"github.com/cuongtranba/faasv/natserver_test"
	"github.com/cuongtranba/faasv/pkgs/request"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const TEST_PORT = 8369

type GatewayServerTestSuite struct {
	suite.Suite
	srv      *GatewayServer
	natQueue faasv.Queue
}

func TestGatewayTestSuite(t *testing.T) {
	natServer := natserver_test.RunNATsServerOnPort(TEST_PORT)
	defer natServer.Shutdown()

	sUrl := fmt.Sprintf("nats://127.0.0.1:%d", TEST_PORT)
	nc, err := nats.Connect(sUrl, nil)
	require.Nil(t, err)
	defer nc.Close()
	natqueue, err := faasv.NewNatQueue(nc)
	require.Nil(t, err)
	defer natqueue.Close()
	go func() {
		natqueue.Start()
	}()
	ctx, canfunc := context.WithCancel(context.Background())
	defer canfunc()
	srv := NewGatewayServer(ctx, natqueue, ":9001")
	defer srv.Stop(30 * time.Second)
	go func() {
		srv.Start()
	}()
	suite.Run(t, &GatewayServerTestSuite{
		natQueue: natqueue,
		srv:      srv,
	})
}

func (g *GatewayServerTestSuite) Test_Should_Call_Success_Http_NatHandler() {
	err := g.natQueue.Subscribe(context.Background(), "test", "worker", func(ctx context.Context, payload string) (string, error) {
		return payload, nil
	})
	require.Nil(g.T(), err)

	res, err := request.Post[Request, Response]("http://localhost:9001", Request{
		Subject: "test",
		Payload: "test",
	})
	require.Nil(g.T(), err)
	time.Sleep(3 * time.Second)
	require.Equal(g.T(), "test", res.Payload.(string))
}

func (g *GatewayServerTestSuite) Test_Should_Call_Success_Socket_NatHandler() {
	err := g.natQueue.Subscribe(context.Background(), "test", "worker", func(ctx context.Context, payload string) (string, error) {
		return payload, nil
	})
	require.Nil(g.T(), err)

	res, err := request.WS[Request, Response]("ws://localhost:9001/ws", Request{
		Subject: "test",
		Payload: "test",
	})
	require.Nil(g.T(), err)
	time.Sleep(3 * time.Second)

	require.Equal(g.T(), "test", res.Payload.(string))
}
