package db

import (
	cache "github.com/zpatrick/go-cache"
)

// KeyValueStoreAdapter allows the use of a store.Store as a slackbot.KeyValueStore
type KeyValueStoreAdapter struct {
	Store
	cache *cache.Cache
	key   string
}

// NewKeyValueStoreAdapter will read/write data to the specified store at the specified key.
// Entries are cached, so Invalidate() must be called prior to Read operations to get the latest data.
func NewKeyValueStoreAdapter(store Store, key string) *KeyValueStoreAdapter {
	return &KeyValueStoreAdapter{
		Store: store,
		cache: cache.New(),
		key:   key,
	}
}

// Invalidate will invalidate the adapter's cache
func (s *KeyValueStoreAdapter) Invalidate() {
	s.cache.Clear()
}

// ReadKeyValues is used to satisfy the slackbot.KeyValueStore interface.
func (s *KeyValueStoreAdapter) ReadKeyValues() (map[string]string, error) {
	if v, ok := s.cache.Getf(s.key); ok {
		return v.(map[string]string), nil
	}

	var kvs map[string]string
	if err := s.Read(s.key, &kvs); err != nil {
		return nil, err
	}

	s.cache.Set(s.key, kvs)
	return kvs, nil
}

// WriteKeyValues is used to satisfy the slackbot.KeyValueStore interface
func (s *KeyValueStoreAdapter) WriteKeyValues(kvs map[string]string) error {
	s.cache.Set(s.key, kvs)
	return s.Write(s.key, kvs)
}
