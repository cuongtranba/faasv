package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cuongtranba/faasv"
	"github.com/cuongtranba/faasv/natserver_test"
	"github.com/cuongtranba/faasv/server"
	"github.com/nats-io/nats.go"
)

const TEST_PORT = 8369

type userRequest struct {
	Query string `json:"query"`
}

func main() {
	natServer := natserver_test.RunNATsServerOnPort(TEST_PORT)
	defer natServer.Shutdown()

	sUrl := fmt.Sprintf("nats://127.0.0.1:%d", TEST_PORT)
	nc, err := nats.Connect(sUrl, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
	natqueue, err := faasv.NewNatQueue(nc)
	if err != nil {
		log.Fatal(err)
	}
	defer natqueue.Close()
	go func() {
		natqueue.Start()
	}()
	ctx := context.Background()
	natqueue.Subscribe(context.Background(), "user.test", "worker", func(ctx context.Context, query userRequest) (string, error) {
		return query.Query, nil
	})

	srv := server.NewGatewayServer(server.WithContext(ctx), server.WithQueue(natqueue), server.WithPort(":8001"))

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		srv.Start()
	}()
	<-stopCh
	srv.Stop(30 * time.Second)
}
