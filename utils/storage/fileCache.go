package storage

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type PersistCache struct {
	mutex        sync.Mutex
	HomeDir      *os.File
	dirPath      string
	fileNameList []string
	layer        ICache
}

var onceC sync.Once
var fileCache *PersistCache

// GetCacheInstance returns the singleton instance of the cache
func NewPersistCache(storagePath string) *PersistCache {
	onceC.Do(func() {
		fileCache = new(PersistCache)
		fileCache.dirPath = storagePath

		// creating the directory if it doesn't exist
		if _, err := os.Stat(fileCache.dirPath); os.IsNotExist(err) {
			os.Mkdir(fileCache.dirPath, 0755)
		}

		// loading the file name mapping
		file, err := os.Open(fileCache.dirPath)
		if err != nil {
			// FileCache: Failed to create the dir
			panic(err)
		}
		fileCache.HomeDir = file

		// getting all the file names
		files, err := file.Readdir(-1)
		if err != nil {
			panic(err)
		}
		for _, file := range files {
			fileCache.fileNameList = append(fileCache.fileNameList, file.Name())
		}
	})
	return fileCache
}

func GetFileCache() *PersistCache {
	if fileCache == nil {
		// failed
		panic("FileCache: trying to get cache instance before it was initlized")
	}
	return fileCache
}

func (f *PersistCache) createFile(fileName string) (*os.File, error) {

	// creating file inside the Working dir
	err := f.HomeDir.Chdir()
	if err != nil {
		return nil, err
	}

	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// The fast cache implements the following interface:
// type ICache interface {
// 	Get(key string) (interface{}, error)
// 	Set(key string, value interface{}) error
// 	Delete(key string) error
// 	SetCacheLayer(layer ICache) error
// 	ActivateLayerSavingRuntime(intervals time.Duration) error
// 	LayerSet(key string, value interface{}) error
// }

func (f *PersistCache) Get(key string) (interface{}, error) {
	// getting the key from the current layer
	f.mutex.Lock()
	defer f.mutex.Unlock()
	result, err := f.readFile(key)
	if err != nil && f.layer != nil {
		// failed to read file, maybe doesn't exists searching in layer
		val, err := f.layer.Get(key)
		if err != nil {
			return nil, err
		}

		err = f.Set(key, val)
		if err != nil {
			// failed to set to the current cache
			return nil, err
		}

		return val, nil
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
func (f *PersistCache) Set(key string, value interface{}) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	// First deleting the file if exsits
	for _, name := range f.fileNameList {
		if name == key {
			err := f.deleteFile(key)
			if err != nil {
				// maybe file was not there
				fmt.Printf("FileCache: Failed to delete file name: %s", key)
			}
			f.fileNameList = removeStringFromList(f.fileNameList, key)
			break
		}
	}

	// Setting new value
	file, err := f.createFile(key)
	if err != nil {
		return fmt.Errorf("FileCache: Error creating file: %v", err)
	}
	defer file.Close()

	val, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("FileCache: Error converting value to []byte")
	}

	_, err = file.Write(val)
	if err != nil {
		return fmt.Errorf("fileCache: Error writing to file: %v", err)
	}

	// adding to the file names
	f.fileNameList = append(f.fileNameList, key)
	return nil
}

func (f *PersistCache) Delete(key string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	err := f.deleteFile(key)
	if err != nil {
		// failed to remove
		return err
	}

	// removing from the file
	f.fileNameList = removeStringFromList(f.fileNameList, key)
	return nil
}

func (f *PersistCache) SetCacheLayer(layer ICache, load bool) error {
	if f.layer != nil {
		return fmt.Errorf("cannot changed layer")
	}
	f.layer = layer
	return nil
}

func (f *PersistCache) ActivateLayerSavingRuntime(intervals time.Duration) error {
	// NONE only to satisfiy the interface
	return nil
}
func (f *PersistCache) LayerSet(key string, value interface{}) error {
	// Like normal Set
	return f.Set(key, value)
}

func (f *PersistCache) GetAllKeys() []string {
	return f.fileNameList
}

func (f *PersistCache) readFile(fileName string) ([]byte, error) {
	err := f.HomeDir.Chdir()
	if err != nil {
		return nil, err
	}
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		// failed to read
		return nil, err
	}

	return data, nil
}

func (f *PersistCache) deleteFile(fileName string) error {
	err := f.HomeDir.Chdir()
	if err != nil {
		return err
	}
	err = os.Remove(fileName)
	if err != nil {
		return err
	}
	return nil
}

func removeStringFromList(slice []string, item string) []string {
	index := -1
	for i, v := range slice {
		if v == item {
			index = i
			break
		}
	}
	if index == -1 {
		return slice
	}
	return append(slice[:index], slice[index+1:]...)
}
