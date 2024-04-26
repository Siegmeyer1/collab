package session

import (
	"errors"
	"sync"
)

type Repository struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
}

func NewRepository() *Repository {
	return &Repository{sessions: make(map[string]*Session)}
}

func (r *Repository) GetSession(roomName string) (*Session, error) {
	s, ok := r.sessions[roomName]

	if !ok {
		return nil, errors.New("no active session")
	}

	return s, nil
}

func (r *Repository) GetOrCreateSession(roomName string) (s *Session, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	s, ok := r.sessions[roomName]

	if !ok {
		s, err = NewSession(roomName)
		if err != nil {
			return nil, err
		}
		r.sessions[roomName] = s
	}

	return s, nil
}
