package abac

import "sync"

type Store interface {
	User(id string) (User, bool)
	Document(id string) (Document, bool)
	UpdateUserPoints(id string, delta int) (int, bool)
}

type MemoryStore struct {
	mu        sync.RWMutex
	users     map[string]User
	documents map[string]Document
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:     map[string]User{},
		documents: map[string]Document{},
	}
}

func (s *MemoryStore) SaveUser(user User) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[user.ID] = user
}

func (s *MemoryStore) SaveDocument(doc Document) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.documents[doc.ID] = doc
}

func (s *MemoryStore) User(id string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[id]
	return user, ok
}

func (s *MemoryStore) Document(id string) (Document, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, ok := s.documents[id]
	return doc, ok
}

func (s *MemoryStore) UpdateUserPoints(id string, delta int) (int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[id]
	if !ok {
		return 0, false
	}

	user.Points += delta
	s.users[id] = user
	return user.Points, true
}
