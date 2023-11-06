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
			fmt.Println("FileCache: Failed to create the dir")
			panic(err)
		}
		fileCache.HomeDir = file

		// getting all the file names
		fileNames, err := file.Readdirnames(0)
		if err != nil {
			// Failed to Get files name
			panic(err)
		}
		fileCache.fileNameList = fileNames

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
	f.fileNameList = append(f.fileNameList, fileName)
	return file, nil
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

	// in fast layer
	return result, nil
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
				fmt.Println("FileCache: Failed to delete file name:", key)
			}
			break
		}
	}

	// Setting new value
	file, err := f.createFile(key)
	if err != nil {
		fmt.Println("FileCache: Error creating file: ", err)
		return err
	}
	defer file.Close()

	val, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("FileCache: Error converting value to []byte")
	}

	_, err = file.Write(val)
	if err != nil {
		fmt.Println("FileCache: Error writing to file: ", err)
		return err
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

func (f *PersistCache) SetCacheLayer(layer ICache, load bool) {
	f.layer = layer
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
