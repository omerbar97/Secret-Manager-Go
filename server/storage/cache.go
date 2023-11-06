package storage

import "time"

// Basic interface for cache implementation
type ICache interface {
	// Getting the key from the cache
	Get(key string) (interface{}, error)

	// Getting all the current keys
	GetAllKeys() []string

	// Setting a new value to the cache
	Set(key string, value interface{}) error

	// Deleting the value from the cache
	Delete(key string) error

	// Applying another cache layer, when the items won't found in the first case
	// then searching in the lower level
	SetCacheLayer(layer ICache, load bool)

	// For each time interval that is set, saving the cache to the lower level
	// updating the lower level if any changes accured
	ActivateLayerSavingRuntime(intervals time.Duration) error

	// Setting the value to the lower cache MAYBE DONT NEED!!!!!
	LayerSet(key string, value interface{}) error
}
