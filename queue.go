package faasv

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type Queue interface {
	Publish(ctx context.Context, subject string, payload any) error
	Subscribe(ctx context.Context, subject, queue string, handler Handler) error
	Request(ctx context.Context, subject string, payload any) (any, error)
	Close() error
	Start()
}

type NatQueue struct {
	natCon    *nats.EncodedConn
	closeChan chan struct{}
}

// Subscribes implements Queue

// Start implements Queue
func (n *NatQueue) Start() {
	<-n.closeChan
}

// Close implements Queue
func (n *NatQueue) Close() error {
	err := n.natCon.Drain()
	if err != nil {
		return err
	}
	n.natCon.Close()
	n.closeChan <- struct{}{}
	return nil
}

// Publish implements Queue
func (n *NatQueue) Publish(ctx context.Context, subject string, payload any) error {
	err := n.natCon.Publish(subject, payload)
	if err != nil {
		return fmt.Errorf("publish subject:%s error: %s", subject, err.Error())
	}
	return nil
}

func (n *NatQueue) Request(ctx context.Context, subject string, payload any) (response any, err error) {
	err = n.natCon.RequestWithContext(ctx, subject, payload, &response)
	if err != nil {
		return nil, fmt.Errorf("request subject:%s error: %s", subject, err.Error())
	}
	return
}

func Request[T, V any](ctx context.Context, queue Queue, subject string, req T) (*V, error) {
	res, err := queue.Request(ctx, subject, req)
	if err != nil {
		return nil, err
	}
	v, ok := res.(V)
	if !ok {
		return nil, fmt.Errorf("request subject:%s error: %s", subject, "invalid response")
	}
	return &v, nil
}

func (n *NatQueue) Subscribe(ctx context.Context, subject, queue string, handler Handler) error {
	errMsg := "subscribe subject:%s error: %s"
	err := isValidHandler(handler)
	if err != nil {
		return fmt.Errorf(errMsg, subject, err.Error())
	}
	_, err = n.natCon.QueueSubscribe(subject, queue, func(subject, reply string, o any) {
		result, err := callHandler(handler, ctx, o)
		if err != nil {
			errReply := n.natCon.Publish(reply, err.Error())
			if errReply != nil {
				log.Error().Msgf(errMsg, subject, errReply.Error())
			}
		}
		if reply != "" {
			errReply := n.natCon.Publish(reply, result)
			if errReply != nil {
				log.Error().Msgf(errMsg, subject, errReply.Error())
			}
		}
	})
	if err != nil {
		return fmt.Errorf(errMsg, subject, err.Error())
	}
	return n.natCon.Flush()
}

func NewNatQueue(conn *nats.Conn) (Queue, error) {
	enc, err := nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	if err != nil {
		return nil, fmt.Errorf("new encoded conn error: %s", err.Error())
	}

	return &NatQueue{
		natCon:    enc,
		closeChan: make(chan struct{}),
	}, nil
}
