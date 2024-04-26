package redis

import (
	"context"
	"diploma/src/logging"
	"diploma/src/utils"
	"encoding/json"
	"github.com/go-redis/redis/v8"
)

var DefaultQueue *Queue

type Queue struct {
	publisherID string
	client      *redis.Client
	pubsub      *redis.PubSub
	subs        map[string]chan<- []byte
}

func NewQueue(client *redis.Client) *Queue {
	return &Queue{
		publisherID: utils.NewPublisherID(),
		client:      client,
		pubsub:      client.Subscribe(context.Background()),
		subs:        make(map[string]chan<- []byte, 0),
	}
}

func (q *Queue) Serve(ctx context.Context) {
	ch := q.pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return

		case message := <-ch:
			subCh, ok := q.subs[message.Channel]
			if !ok {
				continue
			}

			var m queueMessage
			if err := json.Unmarshal([]byte(message.Payload), &m); err != nil {
				logging.Error("unmarshalling queue message from %s: %v", message.Channel, err)
				continue
			}

			if m.PublisherID == q.publisherID {
				continue // skip own messages
			}

			subCh <- m.Message
		}
	}
}

func (q *Queue) Subscribe(ctx context.Context, channel string, subCh chan<- []byte) error {
	if _, ok := q.subs[channel]; ok {
		return nil // already subscribed
	}

	q.subs[channel] = subCh

	if err := q.pubsub.Subscribe(ctx, channel); err != nil {
		return err
	}

	return nil
}

func (q *Queue) Publish(ctx context.Context, channel string, msg []byte) error {
	queueMsg := queueMessage{PublisherID: q.publisherID, Message: msg}
	b, err := json.Marshal(&queueMsg)
	if err != nil {
		return err
	}

	return q.client.Publish(ctx, channel, b).Err()
}

type queueMessage struct {
	PublisherID string `json:"publisher_id"`
	Message     []byte `json:"message"`
}

func (m *queueMessage) Bytes() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	return b
}
