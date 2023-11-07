package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

type FastCache struct {
	mutex    sync.Mutex
	instance *cache.Cache
	layer    ICache
	changed  map[string]bool
}

var once sync.Once
var fastCache *FastCache = nil

func NewFastCache(expirationTime time.Duration) *FastCache {
	// creating the go-cache instance
	if fastCache == nil {
		once.Do(func() {
			changed := make(map[string]bool)
			instance := cache.New(expirationTime, expirationTime*2)
			fastCache = &FastCache{
				instance: instance,
				changed:  changed,
			}
		})
	}
	return fastCache
}

func GetCacheInstance() ICache {
	if fastCache == nil {
		panic("trying to retrive cache instance before it was initlizied!")
	}
	return fastCache
}

// The fast cache implements the following interface:
// type ICache interface {
// 	Get(key string) (interface{}, error)
// 	Set(key string, value interface{}) error
// 	Delete(key string) error
// 	SetCacheLayer(layer ICache)
// 	ActivateLayerSavingRuntime(intervals time.Duration) error
// 	LayerSet(key string, value interface{}) error
// }

func (f *FastCache) GetAllKeys() []string {
	var list []string
	for key := range f.changed {
		list = append(list, key)
	}
	return list
}

func (f *FastCache) Get(key string) (interface{}, error) {
	// getting the key from the current layer
	f.mutex.Lock()
	defer f.mutex.Unlock()
	result, found := f.instance.Get(key)
	if !found && f.layer != nil {
		// was not found in the cache searching in the other layer
		r, err := f.layer.Get(key)
		if err != nil {
			// file was not found in the other layer
			return nil, err
		} else {
			// found in the other layer, applying it to the fast layer
			f.safeSet(key, result)
			f.changed[key] = false
			return r, nil
		}
	}

	// searching if the key is in the fast cache keys
	for _, keyVal := range f.GetAllKeys() {
		if keyVal == key {
			return result, nil
		}
	}
	// in fast layer
	return nil, fmt.Errorf("key found in cache but was deleted before so")
}

func (f *FastCache) ActivateLayerSavingRuntime(intervals time.Duration) error {
	// TODO setting go runtime to save every interval the fast cache to the second cache
	return nil
}

func (f *FastCache) LayerSet(key string, value interface{}) error {
	// Setting the value first in the current cache
	err := f.Set(key, value)
	if err != nil {
		// failed to save to the current level
		return err
	}
	if f.layer != nil {
		err := f.layer.LayerSet(key, value)
		return err
	}
	// layer is nil, dont have where to save it
	return nil
}

func (f *FastCache) Set(key string, value interface{}) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	// setting the value to the fast cache
	f.instance.Set(key, value, 0)
	f.changed[key] = true
	return nil
}

func (f *FastCache) safeSet(key string, value interface{}) error {
	f.instance.Set(key, value, 0)
	return nil
}

func (f *FastCache) Delete(key string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	// for some reason thie Delete won't really delete from the cache
	f.instance.Delete(key)
	delete(f.changed, key)
	return nil
}

func (f *FastCache) SetCacheLayer(layer ICache, load bool) {
	f.layer = layer

	// if load == true then loading the lower cache to the upper cache
	if load {
		lst := f.GetAllKeys()
		for _, key := range lst {
			item, err := f.layer.Get(key)
			if err != nil {
				// failed to get key..
				fmt.Println("failed to Get key:", key)
				continue
			}
			err = f.Set(key, item)
			if err != nil {
				fmt.Println("failed to Set key:", key)
			}
		}
	}
}
