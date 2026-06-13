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
		users: map[string]User{
			"u1": {ID: "u1", Name: "Alice", Department: "legal", Region: "CN", Points: 100},
			"u2": {ID: "u2", Name: "Bob", Department: "legal", Region: "US", Points: 30},
			"u3": {ID: "u3", Name: "Carol", Department: "finance", Region: "CN", Points: 5},
		},
		documents: map[string]Document{
			"doc1": {
				ID:         "doc1",
				Title:      "Legal Draft",
				OwnerID:    "u1",
				Department: "legal",
				Status:     StatusDraft,
				AllowedUsers: map[string]Permission{
					"u2": {CanView: true, CanEdit: false},
				},
				AllowedRegions: map[string]bool{"CN": true},
				MinPoints:      10,
				StartHour:      9,
				EndHour:        18,
			},
			"doc2": {
				ID:         "doc2",
				Title:      "Finance Report",
				OwnerID:    "u3",
				Department: "finance",
				Status:     StatusPublished,
				AllowedUsers: map[string]Permission{
					"u1": {CanView: true, CanEdit: true},
				},
				AllowedRegions: map[string]bool{"CN": true, "US": true},
				MinPoints:      0,
				StartHour:      0,
				EndHour:        24,
			},
		},
	}
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
