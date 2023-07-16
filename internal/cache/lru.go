package cache

import (
	"errors"
	"sync"

	"github.com/mojixcoder/caster/internal/app"
	"github.com/mojixcoder/caster/internal/cache/list"
)

// LRUCache is the LRU cache.
type LRUCache struct {
	mutex    *sync.Mutex
	list     *list.DoublyLinkedList
	storage  map[string]*list.Node
	capacity uint64
}

// Verifying interface compliance.
var _ Cache = (*LRUCache)(nil)

// ErrNotFound is raised when the given key is not found.
var ErrNotFound error = errors.New("not found")

// NewLRUCache returns a new LRU cache.
func NewLRUCache() *LRUCache {
	cache := LRUCache{
		mutex:    new(sync.Mutex),
		list:     list.NewDoublyLinkedList(),
		storage:  make(map[string]*list.Node),
		capacity: app.App.Config.Caster.Capacity,
	}

	return &cache
}

// Get fetches a key from cache.
func (c LRUCache) Get(key string) (any, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if node, ok := c.storage[key]; !ok {
		return nil, ErrNotFound
	} else {
		c.list.MoveToBack(node)
		val := node.GetVal()
		return val, nil
	}
}

// Set sets or overwrites the key-value to cache.
func (c LRUCache) Set(key string, val any) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if node, ok := c.storage[key]; ok {
		node.SetVal(val)
		c.list.MoveToBack(node)
	} else {
		if c.capacity == c.list.Size() {
			key := c.list.RemoveHead()
			delete(c.storage, key)
		}

		node := c.list.AddToBack(key, val)
		c.storage[key] = node
	}

	return nil
}

// Flush resets the cache.
func (c *LRUCache) Flush() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.storage = make(map[string]*list.Node)
	c.list = list.NewDoublyLinkedList()

	return nil
}
