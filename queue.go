package faasv

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type Queue interface {
	Publish(ctx context.Context, subject string, payload any) error
	Subscribe(ctx context.Context, subject string, handler Handler) error
	Request(ctx context.Context, subject string, payload any) (any, error)
	Close() error
	Start()
}

type NatQueue struct {
	jsonPub   *nats.EncodedConn
	closeChan chan struct{}
}

// Subscribes implements Queue

// Start implements Queue
func (n *NatQueue) Start() {
	<-n.closeChan
}

// Close implements Queue
func (n *NatQueue) Close() error {
	err := n.jsonPub.Drain()
	if err != nil {
		return err
	}
	n.jsonPub.Close()
	n.closeChan <- struct{}{}
	return nil
}

// Publish implements Queue
func (n *NatQueue) Publish(ctx context.Context, subject string, payload any) error {
	err := n.jsonPub.Publish(subject, payload)
	if err != nil {
		return fmt.Errorf("publish subject:%s error: %s", subject, err.Error())
	}
	return nil
}

// Request implements Queue
func (n *NatQueue) Request(ctx context.Context, subject string, payload any) (any, error) {
	result := new(any)
	err := n.jsonPub.RequestWithContext(ctx, subject, payload, result)
	if err != nil {
		return nil, fmt.Errorf("request subject:%s - error: %s", subject, err.Error())
	}
	return result, nil
}

func RequestType[T, V any](Queue Queue, ctx context.Context, subject string, payload any) (*V, error) {
	result, err := Queue.Request(ctx, subject, payload)
	var v V
	if err != nil {
		return &v, err
	}
	return result.(*V), nil
}

// Subscribes implements Queue
func (n *NatQueue) Subscribe(ctx context.Context, subject string, handler Handler) error {
	_, err := n.jsonPub.QueueSubscribe(subject, "worker", func(subject, reply string, o any) {
		result, err := callHandler(handler, ctx, o)
		if err != nil {
			log.Err(err).Msgf("handler subject:%s error: %s", subject, err.Error())
		} else {
			err := n.Publish(ctx, reply, result)
			if err != nil {
				log.Err(err).Msgf("publish reply subject:%s, reply:%s error: %s", reply, reply, err.Error())
			}
		}
	})
	if err != nil {
		return fmt.Errorf("subscribe subject:%s error: %s", subject, err.Error())
	}
	return n.jsonPub.Flush()
}

func NewNatQueue(natCon *nats.Conn) (Queue, error) {
	jsonPub, err := nats.NewEncodedConn(natCon, nats.JSON_ENCODER)
	if err != nil {
		return nil, err
	}
	return &NatQueue{
		jsonPub:   jsonPub,
		closeChan: make(chan struct{}),
	}, nil
}
