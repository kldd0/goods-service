package sub

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
)

const (
	clusterID = "dev"
	clientID  = "goods-service"
	channel   = "db-logs"
)

type goodsSubscriber struct {
	sc       stan.Conn
	sub      stan.Subscription
	recvChan chan stan.Msg
}

func New(nc *nats.Conn) (*goodsSubscriber, error) {
	const op = "nats-streaming.sub.New"

	// Connect to NATS cluster
	sc, err := stan.Connect(
		clusterID,
		clientID,
		stan.NatsConn(nc),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatalf("NATS Connection lost, reason: %v", reason)
		}))
	if err != nil {
		return nil, fmt.Errorf("%s: connecting to cluster: %w", op, err)
	}

	log.Printf("Connected to %s clusterID: [%s] clientID: [%s]\n", nats.DefaultURL, clusterID, clientID)

	return &goodsSubscriber{
		sc:       sc,
		recvChan: make(chan stan.Msg),
	}, nil
}

func (s *goodsSubscriber) Subscribe() (recvChan <-chan stan.Msg, err error) {
	const op = "nats-streaming.consumer.Subscribe"

	// Subscribe with manual ack mode
	aw, _ := time.ParseDuration("60s")
	s.sub, err = s.sc.Subscribe(
		channel,
		func(msg *stan.Msg) {
			msg.Ack()

			// sending msg into the output channel
			s.recvChan <- *msg
		},
		stan.MaxInflight(25),
		stan.SetManualAckMode(),
		stan.AckWait(aw),
	)

	if err != nil {
		return nil, fmt.Errorf("%s: subscribing to a channel: %w", op, err)
	}

	log.Printf("Subscribed to the channel: [%s] clientID: [%s]\n", channel, clientID)

	return s.recvChan, nil
}

func (s *goodsSubscriber) Close() {
	if s.sc != nil {
		s.sc.Close()
	}
	if s.sub != nil {
		if err := s.sub.Unsubscribe(); err != nil {
			log.Fatal(err)
		}
	}
	close(s.recvChan)
}
