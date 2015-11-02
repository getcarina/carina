package main

import (
	"fmt"
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

	filename := fmt.Sprintf("carina-temp-cache-%s.json", randomName())

	// Try to clean up as best we can
	defer func() {
		err := os.Remove(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to remove temporary cache: %v\n", err)
		}
	}()

	cache, err := LoadCache(filename)

	if err != nil {
		t.Errorf("Expected nil, got %v\n", err)
	}
	if filename != cache.filename {
		t.Errorf("Expected %v, got %v\n", filename, cache.filename)
	}

	updateTime := time.Now()

	err = cache.UpdateLastCheck(updateTime)
	if err != nil {
		t.Fail()
	}
	newCache, err := LoadCache(filename)
	if err != nil {
		t.Fail()
	}
	if !updateTime.Equal(newCache.LastUpdateCheck) {
		t.Errorf("Expected %v, got %v\n", updateTime, newCache.LastUpdateCheck)
	}

}
