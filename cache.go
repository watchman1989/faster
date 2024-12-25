package faster

import (
	"container/list"
	"time"
)

const (
	ModeLRU  = "LRU"
	ModeFIFO = "FIFO"

	DataTypeKV   = "KV"
	DataTypeHash = "Hash"
)

type entry struct {
	expire   time.Duration
	key      string
	value    interface{}
	dataType string
	hashMap  map[string]interface{}
}

type Evit func(key string, value interface{})

type fastCache struct {
	// LRU, FIFO
	mode      string
	size      int
	evictList *list.List
	dataMap   map[string]*list.Element
	onEvict   Evit
}

func NewFastCache(mode string, size int, onEvict Evit) *fastCache {
	if size <= 0 {
		return nil
	}
	if mode != "LRU" && mode != "FIFO" {
		return nil
	}
	fc := &fastCache{
		mode:      mode,
		size:      size,
		evictList: list.New(),
		dataMap:   make(map[string]*list.Element),
		onEvict:   onEvict,
	}
	return fc
}

func (fc *fastCache) Set(key string, value interface{}, expire time.Duration) {

}
