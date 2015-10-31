package main

import (
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

var rand uint32
var randmu sync.Mutex

func reseed() uint32 {
	return uint32(time.Now().UnixNano() + int64(os.Getpid()))
}

func randomName() string {
	randmu.Lock()
	r := rand
	if r == 0 {
		r = reseed()
	}
	r = r*1664525 + 1013904223 // constants from Numerical Recipes
	rand = r
	randmu.Unlock()
	return strconv.Itoa(int(1e9 + r%1e9))[1:]
}

func TestLoadCache(t *testing.T) {

	noCache := randomName()

	cache, err := LoadCache(noCache)
	if err != ErrCacheNotExist {
		t.Errorf("Expected ErrCacheNotExist, got %v\n", err)
	}
	if cache.filename != noCache {
		t.Errorf("Expected %v, got %v\n", noCache, cache.filename)
	}

	filename := noCache

	cache, err = LoadCache(filename)

	if err != ErrCacheNotExist {
		t.Errorf("Expected ErrCacheNotExist, got %v\n", err)
	}
	if cache.filename != filename {
		t.Errorf("Expected %v, got %v\n", filename, cache.filename)
	}

	updateTime := time.Now()

	cache.updateLastCheck(updateTime)
	newCache, err := LoadCache(filename)
	if newCache.LastUpdateCheck != updateTime {
		t.Errorf("Expected %v, got %v\n", updateTime, newCache.LastUpdateCheck)
	}

}
