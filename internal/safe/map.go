package safe

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
)

// Map is a thread-safe map.
type Map[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

// NewSafeMap creates a new SafeMap.
func NewSafeMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		m: make(map[K]V),
	}
}

// Get returns the value for the given key.
func (sm *Map[K, V]) Get(key K) (*V, bool) {
	log.Debugf("getting value for key: %v", key)
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	val, ok := sm.m[key]
	return &val, ok
}

// Set sets the value for the given key.
func (sm *Map[K, V]) Set(key K, value V) {
	log.Debugf("setting value (%s) for key: %v", reflect.TypeOf(value), key)
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.m[key] = value
}

// Delete deletes the value for the given key.
func (sm *Map[K, V]) Delete(key K) {
	log.Debugf("deleting value (%s) for key: %v", reflect.TypeOf(sm.m), key)
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.m, key)
}

// Len returns the length of the map.
func (sm *Map[K, V]) Len() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.m)
}

// Clear clears the map.
func (sm *Map[K, V]) Clear() {
	log.Debugf("clearing map over type: %s", reflect.TypeOf(sm.m))
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.m = make(map[K]V)
}

// String returns a string representation of the map.
func (sm *Map[K, V]) String() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	var b strings.Builder
	for k, v := range sm.m {
		value := fmt.Sprintf("%v", v)
		value = limitString(value, 100)
		b.WriteString(fmt.Sprintf("%v: %v\n", k, value))
	}
	return b.String()
}

// limitString limits the length of a string to the given limit.
func limitString(s string, limit int) string {
	if len(s) > limit {
		return s[:limit] + "..."
	}
	return s
}

// Values returns the values of the map.
func (sm *Map[K, V]) Values() []V {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	var values []V
	for _, v := range sm.m {
		values = append(values, v)
	}
	return values
}
