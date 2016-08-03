package client

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Cache keeps track of last-check, also intended to cache tokens in the future
// Not all methods are thread safe and it does not use any file locks
type Cache struct {
	sync.Mutex
	filename        string
	LastUpdateCheck time.Time               `json:"last-check"`
	Accounts        map[string]AccountCache `json:"accounts"`
}

func NewCache(filename string) *Cache {
	return &Cache{filename: filename, Accounts: make(map[string]AccountCache)}
}

type AccountCache struct {
	Tokens map[string]string `json:"tokens"`
}

type CacheUnavailableError struct {
	cause error
}

func (error CacheUnavailableError) Error() string {
	return fmt.Sprintf("The cache has been disabled due to the following error: \n%s", error.cause.Error())
}

func (error CacheUnavailableError) Cause() error {
	return error.cause
}

func defaultCacheFilename() (string, error) {
	bd, err := GetCredentialsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(bd, "cache.json"), nil
}

// Read the on disk cache
func (cache *Cache) read() error {
	f, err := os.OpenFile(cache.filename, os.O_RDONLY, 0666)
	if os.IsNotExist(err) {
		return nil
	}

	err = json.NewDecoder(f).Decode(cache)
	if err != nil {
		_ = f.Close()
		return errors.Wrap(err, "Unable to deserialize cache file")
	}
	return f.Close()
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
func LoadCache(filename string) (*Cache, error) {
	cache := NewCache(filename)
	err := cache.read()
	return cache, err
}

// LoadCache retrieves the on disk cache and returns a cache struct
func NilCache() *Cache {
	return NewCache("")
}

// UpdateLastCheck sets the last update time to t
func (cache *Cache) UpdateLastCheck(t time.Time) error {
	if cache.filename == "" {
		return nil
	}

	cache.Lock()
	defer cache.Unlock()

	err := cache.read()
	if err != nil {
		return err
	}
	cache.LastUpdateCheck = t
	return cache.write()
}

// UpdateAccount sets the API token for an account in the cache
func (cache *Cache) UpdateFromAccount(account Account) error {
	if cache.filename == "" {
		return nil
	}

	token := account.Credentials.GetToken()
	if token == "" {
		return nil
	}

	cache.Lock()
	defer cache.Unlock()

	err := cache.read()
	if err != nil {
		return err
	}

	tag := account.GetTag()
	accountCache, exists := cache.Accounts[tag]
	if !exists {
		accountCache = AccountCache{Tokens: make(map[string]string)}
		cache.Accounts[tag] = accountCache
	}

	accountCache.Tokens[account.Credentials.GetUserName()] = token

	return cache.write()
}
