package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/getcarina/carina/common"
	"github.com/pkg/errors"
)

type cacheItem map[string]string

// Cache is an on-disk cache of transient application values
type Cache struct {
	sync.Mutex
	path            string
	LastUpdateCheck time.Time            `json:"last-check"`
	Accounts        map[string]cacheItem `json:"accounts"`
}

// CacheUnavailableError explains why the on-disk cache is unavailable
type CacheUnavailableError struct {
	cause error
}

// Error returns the underlying error message
func (error CacheUnavailableError) Error() string {
	return fmt.Sprintf("The cache has been disabled due to the following error: \n%s", error.cause.Error())
}

// Cause returns the underlying cause of the error
func (error CacheUnavailableError) Cause() error {
	return error.cause
}

func newCache(path string) *Cache {
	return &Cache{
		path:     path,
		Accounts: make(map[string]cacheItem),
	}
}

func defaultCacheFilename() (string, error) {
	bd, err := GetCredentialsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(bd, "cache.json"), nil
}

func (cache *Cache) isNil() bool {
	return cache.path == ""
}

// Load reads the on disk cache into memory
func (cache *Cache) load() error {
	f, err := os.OpenFile(cache.path, os.O_RDONLY, 0666)
	if os.IsNotExist(err) {
		return nil
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(cache)
	if err != nil {
		common.Log.WriteDebug(errors.Wrap(err, "Unable to deserialize cache file, starting over with a fresh cache").Error())
	}

	return nil
}

// Save writes the in memory cache to disk
func (cache *Cache) save() error {
	f, err := os.OpenFile(cache.path, os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		return errors.Wrap(err, "Cannot open on-disk cache")
	}

	err = json.NewEncoder(f).Encode(cache)
	if err != nil {
		_ = f.Close()
		return errors.Wrap(err, "Cannot serialize in-memory cache")
	}

	return nil
}

// update handles locking and loading the on-disk cache before an update
func (cache *Cache) safeUpdate(action func(*Cache)) error {
	if cache.isNil() {
		return nil
	}

	cache.Lock()
	defer cache.Unlock()
	err := cache.load()
	if err != nil {
		return err
	}

	action(cache)

	return cache.save()
}

func (cache *Cache) apply(account Account) {
	accountCache, exists := cache.Accounts[account.GetID()]
	if !exists {
		return
	}

	account.ApplyCache(accountCache)
}

// SaveLastUpdateCheck caches the last time that we checked for updates
func (cache *Cache) SaveLastUpdateCheck(timestamp time.Time) error {
	return cache.safeUpdate(func(c *Cache) {
		c.LastUpdateCheck = timestamp
	})
}

// SaveAccount caches transient account data, such as the auth token
func (cache *Cache) SaveAccount(account Account) error {
	accountCache := account.BuildCache()
	if account == nil {
		common.Log.WriteDebug("Skipping updating the account cache because it is empty")
		return nil
	}

	return cache.safeUpdate(func(c *Cache) {
		c.Accounts[account.GetID()] = accountCache
	})
}
