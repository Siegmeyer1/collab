package app

import (
	"context"
	"diploma/src/session"
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
	sessions *session.Repository
	conn     *websocket.Conn
	incoming chan []byte
}

func NewClient(roomName string, sessions *session.Repository, conn *websocket.Conn) *Client {
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
	defer cancel(nil) // TODO: set cause to "good" close error

	fmt.Printf("client (ID: %s) connected to room %s\n", c.id, c.roomName)

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

		case _ = <-c.incoming:
			//fmt.Println(string(msg))
		}
	}
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

		_, msg, err := c.read(ctx)
		if err != nil {
			return err
		}

		if msg != nil {
			c.incoming <- msg
		}
	}
}
