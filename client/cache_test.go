package client

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/getcarina/carina/common"
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
			common.Log.WriteError("Unable to remove temporary cache: %v", err)
		}
	}()

	cache := newCache(filename)
	err := cache.load()

	if err != nil {
		t.Errorf("Expected nil, got %v\n", err)
	}
	if filename != cache.path {
		t.Errorf("Expected %v, got %v\n", filename, cache.path)
	}

	updateTime := time.Now()

	err = cache.SaveLastUpdateCheck(updateTime)
	if err != nil {
		t.Fail()
	}

	cache = newCache(filename)
	err = cache.load()
	if err != nil {
		t.Fail()
	}
	if !updateTime.Equal(cache.LastUpdateCheck) {
		t.Errorf("Expected %v, got %v\n", updateTime, cache.LastUpdateCheck)
	}

}
