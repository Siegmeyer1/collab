package session

import "sync"

type Repository struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
}

func NewRepository() *Repository {
	return &Repository{sessions: make(map[string]*Session)}
}

func (r *Repository) GetOrCreateSession(roomName string) (*Session, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	s, ok := r.sessions[roomName]

	if !ok {
		s = NewSession(roomName)
		r.sessions[roomName] = s
	}

	return s, nil
}
