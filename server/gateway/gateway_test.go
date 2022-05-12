package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cuongtranba/faasv"
	"github.com/cuongtranba/faasv/natserver_test"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
)

const TEST_PORT = 8369

type userRequest struct {
	Query string `json:"query"`
}

func TestRequestServer(t *testing.T) {
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
	ctx := context.Background()
	err = natqueue.Subscribe(context.Background(), "user.test", "worker", func(ctx context.Context, query userRequest) (string, error) {
		return query.Query, nil
	})
	require.Nil(t, err)
	srv := NewGatewayServer(ctx, natqueue, ":9001")
	defer srv.Stop(30 * time.Second)
	go func() {
		srv.Start()
	}()
	res, err := http.Post("http://localhost:9001/user.test", "application/json", strings.NewReader("{\"query\":\"test\"}"))
	time.Sleep(time.Second * 3)
	require.Nil(t, err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	require.Nil(t, err)
	require.Equal(t, "test", string(body))
}
