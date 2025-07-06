package storage

type Storage interface {
	Get(key string) (*string, error)
	Set(key string, value string) error
}

type InMemoryStorage struct {
	data map[string]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]string),
	}
}

func (s *InMemoryStorage) Get(key string) (*string, error) {
	if value, found := s.data[key]; found {
		return &value, nil
	} else {
		return nil, nil
	}
}

func (s *InMemoryStorage) Set(key string, value string) error {
	s.data[key] = value
	return nil
}

