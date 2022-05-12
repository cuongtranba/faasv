package faasv

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cuongtranba/faasv/natserver_test"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const TEST_PORT = 8369

type QueueTestSuite struct {
	suite.Suite
	natQueue Queue
}

func TestQueueTestSuite(t *testing.T) {
	natServer := natserver_test.RunNATsServerOnPort(TEST_PORT)
	defer natServer.Shutdown()

	sUrl := fmt.Sprintf("nats://127.0.0.1:%d", TEST_PORT)
	nc, err := nats.Connect(sUrl, nil)
	require.Nil(t, err)
	defer nc.Close()
	natqueue, err := NewNatQueue(nc)
	require.Nil(t, err)
	defer natqueue.Close()
	go func() {
		natqueue.Start()
	}()
	suite.Run(t, &QueueTestSuite{
		natQueue: natqueue,
	})
}

func handler(payload string) (string, error) {
	return payload, nil
}

func dosomething(ctx context.Context, userId string) (string, error) {
	return userId, nil
}

func (s *QueueTestSuite) TestPublish() {
	err := s.natQueue.Subscribe(context.Background(), "test", "worker", dosomething)
	require.Nil(s.T(), err)

	err = s.natQueue.Publish(context.Background(), "test", "test")
	time.Sleep(3 * time.Second)
	require.Nil(s.T(), err)
}

func (s *QueueTestSuite) TestRequest() {
	err := s.natQueue.Subscribe(context.Background(), "test", "test001", dosomething)
	require.Nil(s.T(), err)

	res, err := Request[string, string](context.Background(), s.natQueue, "test", "test")
	require.Nil(s.T(), err)
	require.Equal(s.T(), "test", *res)
}
