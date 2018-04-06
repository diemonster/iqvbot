package db

import (
	"github.com/urfave/cli"
	cache "github.com/zpatrick/go-cache"
	"github.com/zpatrick/slackbot"
)

// Keys used for reading & writing data in Stores
const (
	AliasesKey = "aliases"
)

// A StoreAdapter adapts a Store to a slackbot.Store
// Entries are cached, so Invalidate() must be called prior to Read operations to get the latest data
type StoreAdapter struct {
	Store
	cache *cache.Cache
}

func NewStoreAdapter(store Store) *StoreAdapter {
	return &StoreAdapter{
		Store: store,
		cache: cache.New(),
	}
}

// Invalidate will invalidate cache at the given key
func (s *StoreAdapter) Invalidate(key string) {
	s.cache.Delete(key)
}

// InvalidateBefore is a helper function that returns a slackbot.CommandOption
// that invalidates the cache prior to the command executing.
// This should be used on commands that have write-access to the store.
func (s *StoreAdapter) InvalidateBefore(key string) slackbot.CommandOption {
	return slackbot.WithBefore(func(c *cli.Context) error {
		s.Invalidate(key)
		return nil
	})
}

// ReadAliases is used to satisfy the slackbot.AliasStore interface
func (s *StoreAdapter) ReadAliases() (map[string]string, error) {
	if v, ok := s.cache.Getf(AliasesKey); ok {
		return v.(map[string]string), nil
	}

	var aliases map[string]string
	if err := s.Read(AliasesKey, &aliases); err != nil {
		return nil, err
	}

	s.cache.Set(AliasesKey, aliases)
	return aliases, nil
}

// WriteAliases is used to satisfy the slackbot.AliasStore interface
func (s *StoreAdapter) WriteAliases(aliases map[string]string) error {
	s.cache.Set(AliasesKey, aliases)
	return s.Write(AliasesKey, aliases)
}
