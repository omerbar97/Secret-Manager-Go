package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

type FastCache struct {
	mutex        sync.Mutex
	instance     *cache.Cache
	layer        ICache
	changed      map[string]bool
	changedMutex sync.Mutex
	ch           chan bool
}

var once sync.Once
var fastCache *FastCache = nil

func NewFastCache(expirationTime time.Duration) *FastCache {
	// creating the go-cache instance
	if fastCache == nil {
		once.Do(func() {
			chanel := make(chan bool)
			changed := make(map[string]bool)
			instance := cache.New(expirationTime, expirationTime*2)
			fastCache = &FastCache{
				instance: instance,
				changed:  changed,
				ch:       chanel,
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
			f.SetChangedValue(key, false)
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

func (f *FastCache) SetChangedValue(key string, val bool) {
	f.changedMutex.Lock()
	defer f.changedMutex.Unlock()
	f.changed[key] = val
}

func (f *FastCache) ActivateLayerSavingRuntime(intervals time.Duration) error {
	// this function run in different run time, each interval saving to the lower cache the changes
	if f.layer == nil {
		return fmt.Errorf("there is no another layer to the cache")
	}
	go func() {
		for {
			select {
			case <-f.ch:
				fmt.Println("Stopping the Saving Runtime...")
				return
			default:
				// Saving all the changed to the files
				tempChangedKeys := f.changed
				for key, val := range tempChangedKeys {
					if val {
						// this key was changed
						if realVal, err := f.Get(key); err != nil {
							// failed to save realVal
							fmt.Println("Cache Runtime: Failed While trying to get key:", key, "and save it to the lower cache level")
						} else if err := f.layer.Set(key, realVal); err != nil {
							fmt.Println("Cache Runtime: Failed While trying to save key:", key, "to the lower cache level")
						} else {
							fmt.Println("Saving key:", key, "to lower level")
							f.SetChangedValue(key, false)
						}
					}
				}
				// Sleep for the defined interval
				time.Sleep(intervals)
			}
		}
	}()
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
	f.SetChangedValue(key, true)
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

func (f *FastCache) SetCacheLayer(layer ICache, load bool) error {
	if f.layer != nil {
		return fmt.Errorf("cannot changed layer")
	}

	f.layer = layer

	// if load == true then loading the lower cache to the upper cache
	if load {
		lst := f.layer.GetAllKeys()
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
	return nil
}
