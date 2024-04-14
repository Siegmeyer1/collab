package session

import (
	"diploma/src/utils"
)

type Client interface {
	ID() string
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
