package engine

import (
	"path"
	"sync"
)

// BaseEngine provides common functionality for template engines
// It includes path resolution, caching support, and thread-safe operations
type BaseEngine struct {
	Basedir      string       // Template directory
	CacheEnabled bool         // Whether caching is enabled
	Mu           sync.RWMutex // Protects cache operations (exported for testing)
}

// GetFullPath returns the full path for a template file
func (b *BaseEngine) GetFullPath(name string) string {
	return path.Join(b.Basedir, name)
}

// Lock acquires a write lock (for cache writes)
func (b *BaseEngine) Lock() {
	b.Mu.Lock()
}

// Unlock releases a write lock
func (b *BaseEngine) Unlock() {
	b.Mu.Unlock()
}

// RLock acquires a read lock (for cache reads)
func (b *BaseEngine) RLock() {
	b.Mu.RLock()
}

// RUnlock releases a read lock
func (b *BaseEngine) RUnlock() {
	b.Mu.RUnlock()
}

// NewBaseEngine creates a new base engine
func NewBaseEngine(basedir string, cacheEnabled bool) BaseEngine {
	return BaseEngine{
		Basedir:      basedir,
		CacheEnabled: cacheEnabled,
	}
}
