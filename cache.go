package main

import (
	"encoding/json"
	"os"
	"path"
	"sync"
	"time"
)

// Cache keeps track of last-check, also intended to cache tokens in the future
// Not all methods are thread safe and it does not use any file locks
type Cache struct {
	sync.Mutex
	filename        string
	LastUpdateCheck time.Time         `json:"last-check"`
	Tokens          map[string]string `json:"tokens"`
}

func defaultCacheFilename() (string, error) {
	bd, err := CarinaCredentialsBaseDir()
	if err != nil {
		return "", err
	}
	return path.Join(bd, "cache.json"), nil
}

// Read the on disk cache
func (cache *Cache) read() error {
	f, err := os.OpenFile(cache.filename, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	err = json.NewDecoder(f).Decode(cache)
	if err != nil {
		_ = f.Close()
		return err
	}
	err = f.Close()
	return err
}

// write will put the current version of the cache on disk at cache.filename
// creating it if it does not exist
func (cache *Cache) write() error {
	f, err := os.OpenFile(cache.filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	err = json.NewEncoder(f).Encode(cache)
	if err != nil {
		_ = f.Close()
		return err
	}
	err = f.Close()
	return err
}

// LoadCache retrieves the on disk cache and returns a cache struct
func LoadCache(filename string) (cache *Cache, err error) {
	cache = new(Cache)
	cache.Tokens = make(map[string]string)

	cache.filename = filename
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(cache.filename)

	if os.IsNotExist(err) {
		return cache, cache.write()
	} else if err != nil {
		return nil, err
	}

	return cache, cache.read()
}

// UpdateLastCheck sets the last update time to t
func (cache *Cache) UpdateLastCheck(t time.Time) error {
	cache.Lock()
	defer cache.Unlock()

	err := cache.read()
	if err != nil {
		return err
	}
	cache.LastUpdateCheck = t
	return cache.write()
}

// SetToken sets the API token for a username in the cache
func (cache *Cache) SetToken(username, token string) error {
	cache.Lock()
	defer cache.Unlock()

	err := cache.read()
	if err != nil {
		return err
	}

	cache.Tokens[username] = token
	return cache.write()
}
