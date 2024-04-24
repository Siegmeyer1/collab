package session

import (
	"context"
	"diploma/src/logging"
	"diploma/src/messages"
	"diploma/src/utils"
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
	sessions *Repository
	conn     *websocket.Conn

	incoming chan []byte
}

func NewClient(roomName string, sessions *Repository, conn *websocket.Conn) *Client {
	return &Client{
		id:       utils.NewClientID(),
		roomName: roomName,
		sessions: sessions,
		conn:     conn,
		incoming: make(chan []byte, bufferSize),
	}
}

func (c *Client) ID() string {
	return c.id
}

func (c *Client) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(context.Canceled)

	logging.Debug("client (ID: %s) connected to room %s", c.id, c.roomName)

	sess, err := c.sessions.GetOrCreateSession(c.roomName)
	if err != nil {
		return err
	}

	sess.AddClient(c)

	go func() {
		err := c.listen(ctx)
		cancel(err)
	}()

	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)

		case msg := <-c.incoming:
			if err := c.handleMessage(ctx, msg); err != nil {
				return err
			}
		}
	}
}

func (c *Client) Close() error {
	sess, err := c.sessions.GetSession(c.roomName)
	if err != nil {
		return err
	}

	sess.RemoveClient(c)

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
			//msg, err := messages.DecodeMessage(msgBytes)
			//if err != nil {
			//	logging.Error("decoding message: %v", err)
			//	continue
			//}

			c.incoming <- msgBytes
		}
	}
}

func (c *Client) handleMessage(ctx context.Context, bytes []byte) error {
	sess, err := c.sessions.GetSession(c.roomName)
	if err != nil {
		return err
	}

	protocol, messageType, err := messages.PeekProtoAndType(bytes)
	if err != nil {
		return err
	}

	if protocol == messages.AwarenessProtocol {
		go sess.SendMessage(ctx, c, bytes)
		return nil
	}

	if protocol != messages.SyncProtocol {
		return fmt.Errorf("unknown protocol: %d", protocol)
	}

	switch messageType {
	case messages.Update:
		go sess.SendMessage(ctx, c, bytes)

		updateMessage, err := messages.DecodeUpdateMessage(bytes)
		if err != nil {
			return err
		}

		if updateMessage.IsDeleteOnly {
			return sess.removalRepo.StoreRemoval(updateMessage.Data)
		}
		return sess.updateRepo.StoreUpdate(updateMessage)

	case messages.SyncRequest:
		syncMessage, err := messages.DecodeSyncReqMessage(bytes)
		if err != nil {
			return err
		}

		updates, err := sess.updateRepo.GetUpdates(syncMessage)
		if err != nil {
			return err
		}

		for _, update := range updates {
			err := c.Send(ctx, update)
			if err != nil {
				return err
			}
		}

		removals, err := sess.removalRepo.GetRemovals()
		if err != nil {
			return err
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
