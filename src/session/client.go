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

	incoming chan *messages.Message
}

func NewClient(roomName string, sessions *Repository, conn *websocket.Conn) *Client {
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
			protocol, msgType, err := messages.PeekProtoAndType(msgBytes)
			if err != nil {
				logging.Exception(err)
				continue
			}

			if protocol == messages.SyncProtocol {
				switch msgType {
				case messages.Update:
					msg, err := messages.DecodeUpdateMessage(msgBytes)
					if err != nil {
						logging.Error("decoding update: %v", err)
					}
					logging.Debug("got update (ID: %d, clock: %d)", msg.ClientID, msg.Clock)
				case messages.SyncStep1:
					msg, err := messages.DecodeStep1SyncMessage(msgBytes)
					if err != nil {
						logging.Error("decoding sync: %v", err)
					}
					logging.Debug("got sync request (%v)", msg)
				}
			}

			msg, err := messages.DecodeMessage(msgBytes)
			if err != nil {
				logging.Error("decoding message: %v", err)
				continue
			}
			//if msg.Protocol == messages.SyncProtocol && msg.MessageType == messages.SyncStep1 {
			//	msg2, err := messages.DecodeStep1SyncMessage(msg.Data)
			//	if err != nil {
			//		logging.Exception(err)
			//	}
			//	logging.Info("Sync request: %v", msg2)
			//}
			c.incoming <- msg
		}
	}
}

func (c *Client) handleMessage(ctx context.Context, msg *messages.Message) error {
	sess, err := c.sessions.GetSession(c.roomName)
	if err != nil {
		return err
	}

	if msg.Protocol == messages.AwarenessProtocol {
		go sess.SendMessage(ctx, c, msg.Data)
		return nil
	}

	if msg.Protocol != messages.SyncProtocol {
		return fmt.Errorf("unknown protocol: %d", msg.Protocol)
	}

	switch msg.MessageType {
	case messages.Update:
		go sess.SendMessage(ctx, c, msg.Data)

		updateMessage, err := messages.DecodeUpdateMessage(msg.Data)
		if err != nil {
			return err
		}

		if updateMessage.IsDeleteOnly {
			return sess.removalRepo.StoreRemoval(updateMessage.DeleteData)
		}
		return sess.updateRepo.StoreUpdate(updateMessage)

	case messages.SyncStep1:
		syncMessage, err := messages.DecodeStep1SyncMessage(msg.Data)
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
		return fmt.Errorf("unexpected msg type: %d", msg.MessageType)
	}

	// TODO: send concurrently (in goroutine)
	//return sess.SendMessage(ctx, c, msg.Data)
}

//func (c *Client) handle(ctx context.Context) error {
//	for {
//		select {
//		case <-ctx.Done():
//			return ctx.Err()
//		case msg := <-c.got:
//			err := c.handleMessage(ctx, msg)
//			if err != nil {
//				logging.Exception(err)
//			}
//
//		}
//
//	}
//}

func (c *Client) Send(ctx context.Context, msg []byte) error {
	return c.conn.Write(ctx, websocket.MessageBinary, msg)
}
