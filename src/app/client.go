package app

import (
	"context"
	"diploma/src/logging"
	"diploma/src/messages"
	"diploma/src/session"
	"diploma/src/utils"
	"nhooyr.io/websocket"
	"time"
)

const (
	bufferSize = 10
)

type Client struct {
	id       string
	roomName string
	sessions *session.Repository
	conn     *websocket.Conn

	incoming chan *messages.Message
}

func NewClient(roomName string, sessions *session.Repository, conn *websocket.Conn) *Client {
	return &Client{
		id:       utils.NewClientID(),
		roomName: roomName,
		sessions: sessions,
		conn:     conn,
		incoming: make(chan *messages.Message, bufferSize),
		//got:      make(chan *messages.Message, bufferSize),
	}
}

func (c *Client) ID() string {
	return c.id
}

func (c *Client) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil) // TODO: set cause to "good" close error

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

	//go func() {
	//	err := c.handle(ctx)
	//	cancel(err)
	//}()

	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)

		case msg := <-c.incoming:
			if msg.Protocol != messages.AwarenessProtocol {
				logging.Debug("client (%s) sent message (proto: %d | type: %d)", c.id, msg.Protocol, msg.MessageType)
			}
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
			msg, err := messages.DecodeMessage(msgBytes)
			if err != nil {
				logging.Error("decoding message: %v", err)
				continue
			}
			c.incoming <- msg
		}
	}
}

func (c *Client) handleMessage(ctx context.Context, msg *messages.Message) error {
	sess, err := c.sessions.GetSession(c.roomName)
	if err != nil {
		return err
	}
	// TODO: send concurrently (in goroutine)
	return sess.SendAllExcept(ctx, c, msg.Data)
}

//func (c *Client) handle(ctx context.Context) error {
//	for {
//		select {
//		case <-ctx.Done():
//			return ctx.Err()
//		case msg := <-c.got:
//			logging.Debug("message from client %s: %s", c.id, string(msg))
//
//		}
//
//	}
//}

func (c *Client) Send(ctx context.Context, msg []byte) error {
	return c.conn.Write(ctx, websocket.MessageBinary, msg)
}
