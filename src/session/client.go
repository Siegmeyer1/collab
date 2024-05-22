package session

import (
	"collab/src/logging"
	"collab/src/messages"
	"collab/src/utils"
	"context"
	"fmt"
	"nhooyr.io/websocket"
	"time"
)

const (
	bufferSize = 10
)

type Client struct {
	id       string
	roomName string
	sess     *Session
	conn     *websocket.Conn

	incoming chan []byte
}

func NewClient(roomName string, session *Session) *Client {
	return &Client{
		id:       utils.NewClientID(),
		roomName: roomName,
		sess:     session,
		incoming: make(chan []byte, bufferSize),
	}
}

func (c *Client) ID() string {
	return c.id
}

func (c *Client) Start(ctx context.Context, conn *websocket.Conn) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(context.Canceled)

	c.conn = conn

	go func() {
		err := c.listen(ctx)
		cancel(err)
	}()

	c.sess.AddClient(c)

	logging.Debug("client (ID: %s) connected to room %s", c.id, c.roomName)

	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)

		case msg := <-c.incoming:
			msg2 := make([]byte, len(msg))
			if n := copy(msg2, msg); n != len(msg) {
				logging.Error("bad copy")
				continue
			}
			go func(msg []byte) {
				if err := c.handleMessage(ctx, msg); err != nil {
					logging.Exception(err)
				}
			}(msg2)
		}
	}
}

func (c *Client) Close() error {
	c.sess.RemoveClient(c)

	return nil
}

func (c *Client) read(ctx context.Context) (websocket.MessageType, []byte, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Minute*60)
	defer cancel()

	return c.conn.Read(ctxTimeout)
}

func (c *Client) listen(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		wsMsgType, msgBytes, err := c.read(ctx)
		if err != nil {
			return err
		}

		if wsMsgType != websocket.MessageBinary {
			continue
		}

		if msgBytes != nil {
			c.incoming <- msgBytes
		}
	}
}

func (c *Client) handleMessage(ctx context.Context, bytes []byte) error {
	protocol, messageType, err := messages.PeekProtoAndType(bytes)
	if err != nil {
		return err
	}

	if protocol == messages.AwarenessProtocol {
		go c.sess.SendMessage(ctx, c, bytes)
		return nil
	}

	if protocol != messages.SyncProtocol {
		return fmt.Errorf("unknown protocol: %d", protocol)
	}

	switch messageType {
	case messages.Update:
		updateMessage, err := messages.DecodeUpdateMessage(bytes)
		if err != nil {
			return err
		}

		if updateMessage.IsDeleteOnly {
			if err = c.sess.removalRepo.StoreRemoval(updateMessage.Data); err != nil {
				return err
			}
		} else if err = c.sess.updateRepo.StoreUpdate(updateMessage); err != nil {
			return err
		}

		go c.sess.SendMessage(ctx, c, bytes)

		return nil

	case messages.SyncRequest:
		syncMessage, err := messages.DecodeSyncReqMessage(bytes)
		if err != nil {
			return err
		}

		updates, err := c.sess.updateRepo.GetUpdates(syncMessage)
		if err != nil {
			return err
		}

		removals, err := c.sess.removalRepo.GetRemovals()
		if err != nil {
			return err
		}

		for _, update := range updates {
			err := c.Send(ctx, update)
			if err != nil {
				return err
			}
		}

		for _, removal := range removals {
			err := c.Send(ctx, removal)
			if err != nil {
				return err
			}
		}

		return nil

	default:
		return fmt.Errorf("unexpected msg type: %d", messageType)
	}
}

func (c *Client) Send(ctx context.Context, msg []byte) error {
	return c.conn.Write(ctx, websocket.MessageBinary, msg)
}
