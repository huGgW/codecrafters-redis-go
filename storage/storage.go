package storage

import (
	"log/slog"
	"time"

	"github.com/codecrafters-io/redis-starter-go/pkg"
)

type Storage interface {
	Get(key string) (*string, error)
	Set(key string, value string, expireAt *time.Time) error
	ExpireAllUntil(time time.Time)
}

type expirationEntry struct {
	key      string
	expireAt time.Time
}

type InMemoryStorage struct {
	data           map[string]string
	expirationMap  map[string]time.Time
	expirationHeap *pkg.Heap[expirationEntry]
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data:          make(map[string]string),
		expirationMap: make(map[string]time.Time),
		expirationHeap: pkg.NewHeap(func(e1, e2 expirationEntry) bool {
			if e1.expireAt.Before(e2.expireAt) {
				return true
			}

			if e1.expireAt.Equal(e2.expireAt) && e1.key < e2.key {
				return true
			}

			return false
		}),
	}
}

func (s *InMemoryStorage) Get(key string) (*string, error) {
	value, found := s.data[key]
	if !found {
		return nil, nil
	}

	expireAt, found := s.expirationMap[key]
	if found && time.Now().After(expireAt) {
		return nil, nil
	}

	return &value, nil
}

func (s *InMemoryStorage) Set(key string, value string, expireAt *time.Time) error {
	if expireAt != nil {
		if expireAt.Before(time.Now()) {
			s.Delete(key)
			return nil
		}

		pastExpiredAt, found := s.expirationMap[key]
		if !found || pastExpiredAt.After(*expireAt) {
			s.expirationMap[key] = *expireAt
		}

		s.expirationHeap.Push(expirationEntry{
			key:      key,
			expireAt: *expireAt,
		})

	} else {
		delete(s.expirationMap, key)
	}

	s.data[key] = value
	return nil
}

func (s *InMemoryStorage) ExpireAllUntil(time time.Time) {
	for s.expirationHeap.Len() > 0 {
		entry, _ := s.expirationHeap.Peek()
		if !entry.expireAt.Before(time) {
			break
		}

		entry, _ = s.expirationHeap.Pop()
		if latestExpiration, exists := s.expirationMap[entry.key]; exists && latestExpiration.After(time) {
			continue
		}

		s.Delete(entry.key)
		slog.Info("expired key removed",
			slog.String("key", entry.key),
		)
	}
}

func (s *InMemoryStorage) Delete(key string) {
	delete(s.data, key)
	delete(s.expirationMap, key)
}
