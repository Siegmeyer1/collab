package session

import (
	"context"
	"diploma/src/document"
	"diploma/src/logging"
	"diploma/src/postgres"
	"diploma/src/redis"
	"diploma/src/utils"
	"sync"
)

type Session struct {
	id           string
	roomName     string
	clients      map[*Client]struct{}
	clientsMux   sync.RWMutex
	errChan      chan error
	msgQueueCh   chan []byte
	updateRepo   document.UpdateRepository
	removalRepo  document.RemovalRepository
	cancelListen context.CancelFunc
}

func NewSession(roomName string) (*Session, error) {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Session{
		id:           utils.NewSessID(),
		roomName:     roomName,
		clients:      make(map[*Client]struct{}),
		errChan:      make(chan error, 10),
		msgQueueCh:   make(chan []byte, 100),
		updateRepo:   postgres.NewUpdateRepository(roomName),
		removalRepo:  postgres.NewRemovalRepository(roomName),
		cancelListen: cancel,
	}

	if err := redis.DefaultQueue.Subscribe(ctx, roomName, s.msgQueueCh); err != nil {
		return nil, err
	}

	go s.listen(ctx)

	return s, nil
}

func (s *Session) AddClient(c *Client) {
	s.clientsMux.Lock()
	defer s.clientsMux.Unlock()

	s.clients[c] = struct{}{}
}

func (s *Session) RemoveClient(c *Client) {
	s.clientsMux.Lock()
	defer s.clientsMux.Unlock()

	delete(s.clients, c)
}

func (s *Session) SendMessage(ctx context.Context, c *Client, msg []byte) {
	if err := redis.DefaultQueue.Publish(ctx, s.roomName, msg); err != nil {
		logging.Exception(err)
		return
	}

	if err := s.sendMessageLocal(ctx, c, msg); err != nil {
		logging.Exception(err)
	}
}

func (s *Session) sendMessageLocal(ctx context.Context, c *Client, msg []byte) error {
	s.clientsMux.RLock()
	defer s.clientsMux.RUnlock()

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

func (s *Session) listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-s.msgQueueCh:
			if err := s.sendMessageLocal(ctx, nil, msg); err != nil {
				logging.Exception(err)
			}
		}
	}
}
