package session

import (
	"context"
	"diploma/src/utils"
)

type Client interface {
	ID() string
	Send(context.Context, []byte) error
}

type Session struct {
	id       string
	roomName string
	clients  map[Client]struct{}
}

func NewSession(roomName string) *Session {
	s := &Session{
		id:       utils.NewSessID(),
		roomName: roomName,
		clients:  make(map[Client]struct{}),
	}

	return s
}

func (s *Session) AddClient(c Client) {
	s.clients[c] = struct{}{}
}

func (s *Session) RemoveClient(c Client) {
	delete(s.clients, c)
}

func (s *Session) SendAllExcept(ctx context.Context, c Client, msg []byte) error {
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

func (s *Session) SendAll(ctx context.Context, msg []byte) error {
	for client := range s.clients {
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
