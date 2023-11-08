package storage

import (
	"fmt"
	GenericEncoding "golang-secret-manager/utils/genericEncoding"
	"reflect"
	"time"
)

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
	SetCacheLayer(layer ICache, load bool) error

	// For each time interval that is set, saving the cache to the lower level
	// updating the lower level if any changes accured
	ActivateLayerSavingRuntime(intervals time.Duration) error

	// Setting the value to the lower cache MAYBE DONT NEED!!!!!
	LayerSet(key string, value interface{}) error
}

func GetCacheValue[T any](cache ICache, key string) (*T, error) {
	// getting the value from the ICache
	val, err := cache.Get(key)
	if err != nil {
		// couldn't found
		return nil, fmt.Errorf("couldn't found key: %s in cache: %v", key, err)
	}
	// converting the value of the result
	result, err := GenericEncoding.FromJson[T](val)
	if err != nil {
		// couldn't found
		return nil, fmt.Errorf("couldn't convert key: %s in cache found type: %T error: %v", key, reflect.TypeOf(result), err)
	}
	if result != nil {
		return result, nil
	}
	return nil, fmt.Errorf("nil pointer was found in the cache")
}

func SetCacheValue[T any](cache ICache, key string, value T) error {
	bytes, err := GenericEncoding.ToJson[T](value)
	if err != nil {
		return fmt.Errorf("CACHE: failed to cache value: %v", err)
	} else {
		err = cache.Set(key, bytes)
		if err != nil {
			return fmt.Errorf("CACHE: failed to cache value: %v", err)
		}
	}
	return nil
}
