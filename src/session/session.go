package session

import (
	"context"
	"diploma/src/document"
	"diploma/src/postgres"
	"diploma/src/utils"
)

type Session struct {
	id          string
	roomName    string
	clients     map[*Client]struct{}
	updateRepo  document.UpdateRepository
	removalRepo document.RemovalRepository
}

func NewSession(roomName string) *Session {
	s := &Session{
		id:          utils.NewSessID(),
		roomName:    roomName,
		clients:     make(map[*Client]struct{}),
		updateRepo:  postgres.NewUpdateRepository(roomName),
		removalRepo: postgres.NewRemovalRepository(roomName),
	}

	return s
}

func (s *Session) AddClient(c *Client) {
	s.clients[c] = struct{}{}
}

func (s *Session) RemoveClient(c *Client) {
	delete(s.clients, c)
}

func (s *Session) SendMessage(ctx context.Context, c *Client, msg []byte) error {
	return s.sendMessageLocal(ctx, c, msg)
}

func (s *Session) sendMessageLocal(ctx context.Context, c *Client, msg []byte) error {
	for client := range s.clients {
		if client == c {
			continue
		}
		if err := client.Send(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

//func (s *Session) Start(ctx context.Context) error {
//
//
//	return nil
//}
