package main

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"time"
)

// Cache keeps track of last-check, also intended to cache tokens in the future
type Cache struct {
	filename        string
	LastUpdateCheck time.Time `json:"last-check"`
}

func defaultCacheFilename() (string, error) {
	bd, err := CarinaCredentialsBaseDir()
	if err != nil {
		return "", err
	}
	return path.Join(bd, "cache.json"), nil
}

// ErrCacheNotExist informs about if the cache does not exist
var ErrCacheNotExist = errors.New("Cache does not exist")

// write will put the current version of the cache on disk at cache.filename
// creating it if it does not exist
func (cache *Cache) write() error {
	f, err := os.OpenFile(cache.filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	err = json.NewEncoder(f).Encode(cache)
	if err != nil {
		return err
	}
	err = f.Close()
	return err
}

func loadCache(filename string) (cache *Cache, err error) {
	cache = new(Cache)

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

	f, err := os.OpenFile(cache.filename, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(f).Decode(cache)
	if err != nil {
		return nil, err
	}
	err = f.Close()
	return cache, err

}

func (cache *Cache) updateLastCheck(t time.Time) error {
	cache.LastUpdateCheck = t
	return cache.write()
}

func (cache *Cache) shouldCheckLatest() (check bool, err error) {
	lastCheck := cache.LastUpdateCheck

	// If we last checked `delay` ago, don't check again
	if lastCheck.Add(12 * time.Hour).After(time.Now()) {
		return false, nil
	}

	err = cache.updateLastCheck(time.Now())
	return true, err
}
