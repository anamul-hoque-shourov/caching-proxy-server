package cache

import (
	"sync"
	"time"
)

type Item struct {
	Data      []byte
	ExpiresAt time.Time
}

type Cache struct {
	items map[string]Item
	mu    sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		items: make(map[string]Item),
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	if !found {
		return nil, false
	}

	if time.Now().After(item.ExpiresAt) {
		return nil, false
	}

	return item.Data, true
}

func (c *Cache) Set(key string, data []byte, ttl time.Duration) {
	c.mu.Lock()
	c.items[key] = Item{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
	}
	c.mu.Unlock()
}
