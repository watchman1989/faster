package faster

import (
	crand "crypto/rand"
	"math"
	"math/big"
	mrand "math/rand"
	"os"
	"sync"
	"time"
)

type Faster struct {
	seed   uint32
	num    uint32
	mus    []sync.Mutex
	shards []*fasterCache
}


func NewFaster(mode int, num uint32, size int, onEvict EvictFunc) *Faster {
	//generate a seed, used for djb33
	var seed uint32
	max := big.NewInt(0).SetUint64(uint64(math.MaxUint32))
	rnd, err := crand.Int(crand.Reader, max)
	if err != nil {
		_, _ = os.Stderr.Write([]byte("\n"))
		seed = mrand.Uint32()
	} else {
		seed = uint32(rnd.Uint64())
	}
	//
	hc := &Faster{
		seed:   seed,
		num:    num,
		mus:    make([]sync.Mutex, num),
		shards: make([]*fasterCache, num),
	}
	//init HLru
	for i := uint32(0); i < num; i++ {
		hc.shards[i] = NewFasterCache(mode, size, onEvict)
	}
	return hc
}


func (f *Faster) idx(k string) uint32 {
	return djb33(f.seed, k) % f.num
}


func (f *Faster) Set(key string, value interface{}, expiration time.Duration) {
	i := f.idx(key)
	f.mus[i].Lock()
	defer f.mus[i].Unlock()
	f.shards[i].set(key, value, expiration)
}

func (f *Faster) Get(key string) interface{} {
	i := f.idx(key)
	f.mus[i].Lock()
	defer f.mus[i].Unlock()
	return f.shards[i].get(key)
}


func (f *Faster) DataType(key string) int {
	i := f.idx(key)
	f.mus[i].Lock()
	defer f.mus[i].Unlock()
	return f.shards[i].dataType(key)
}

func (f *Faster) Del(key string) {
	i := f.idx(key)
	f.mus[i].Lock()
	defer f.mus[i].Unlock()
	f.shards[i].del(key)
}

func (f *Faster) Exist(key string) bool {
	i := f.idx(key)
	f.mus[i].Lock()
	defer f.mus[i].Unlock()
	return f.shards[i].exist(key)
}

func (f *Faster) Len() int {
	tLen := 0
	for i := 0; i < int(f.num); i++ {
		f.mus[i].Lock()
		tLen += f.shards[i].len()
		f.mus[i].Unlock()
	}
	return tLen
}

func (f *Faster) Keys() []string {
	keys := make([]string, 0)
	for i := 0; i < int(f.num); i++ {
		f.mus[i].Lock()
		sKeys := f.shards[i].keys()
		keys = append(keys, sKeys...)
		f.mus[i].Unlock()
	}
	return keys
}

func (f *Faster) IncrBy(key string, incr int64) int64 {
	i := f.idx(key)
	f.mus[i].Lock()
	defer f.mus[i].Unlock()
	return f.shards[i].incrBy(key, incr)
}


func (f *Faster) GetTTL(key string) int64 {
	i := f.idx(key)
	f.mus[i].Lock()
	defer f.mus[i].Unlock()
	return f.shards[i].getTTL(key)
}


func (f *Faster) Expire(key string, expiration time.Duration) {
	i := f.idx(key)
	f.mus[i].Lock()
	defer f.mus[i].Unlock()
	f.shards[i].expire(key, expiration)
}

func (f *Faster) HSet(key, subKey string, value interface{}, expiration time.Duration) {
	i := f.idx(key)
	f.mus[i].Lock()
	defer f.mus[i].Unlock()
	f.shards[i].hSet(key, subKey, value, expiration)
}

func (f *Faster) HGet(key, subKey string) interface{} {
	i := f.idx(key)
	f.mus[i].Lock()
	defer f.mus[i].Unlock()
	return f.shards[i].hGet(key, subKey)
}

func (f *Faster) HExist(key, subKey string) bool {
	i := f.idx(key)
	f.mus[i].Lock()
	defer f.mus[i].Unlock()
	return f.shards[i].hExist(key, subKey)
}


func (f *Faster) HDel(key, subKey string) {
	i := f.idx(key)
	f.mus[i].Lock()
	defer f.mus[i].Unlock()
	f.shards[i].hDel(key, subKey)
}

func (f *Faster) HGetAll(key string) map[string]interface{} {
	i := f.idx(key)
	f.mus[i].Lock()
	defer f.mus[i].Unlock()
	return f.shards[i].hGetAll(key)
}


func (f *Faster) HLen(key string) int {
	i := f.idx(key)
	f.mus[i].Lock()
	defer f.mus[i].Unlock()
	return f.shards[i].hLen(key)
}


func (f *Faster) HKeys(key string) []string {
	i := f.idx(key)
	f.mus[i].Lock()
	defer f.mus[i].Unlock()
	return f.shards[i].hKeys(key)
}

func (f *Faster) HIncrBy(key, subKey string, incr int64) int64 {
	i := f.idx(key)
	f.mus[i].Lock()
	defer f.mus[i].Unlock()
	return f.shards[i].hIncrBy(key, subKey, incr)
}




// djb2 with better shuffling. 5x faster than FNV with the hash.Hash overhead.
func djb33(seed uint32, k string) uint32 {
	var (
		l = uint32(len(k))
		d = 5381 + seed + l
		i = uint32(0)
	)
	// Why is all this 5x faster than a for loop?
	if l >= 4 {
		for i < l-4 {
			d = (d * 33) ^ uint32(k[i])
			d = (d * 33) ^ uint32(k[i+1])
			d = (d * 33) ^ uint32(k[i+2])
			d = (d * 33) ^ uint32(k[i+3])
			i += 4
		}
	}
	switch l - i {
	case 1:
	case 2:
		d = (d * 33) ^ uint32(k[i])
	case 3:
		d = (d * 33) ^ uint32(k[i])
		d = (d * 33) ^ uint32(k[i+1])
	case 4:
		d = (d * 33) ^ uint32(k[i])
		d = (d * 33) ^ uint32(k[i+1])
		d = (d * 33) ^ uint32(k[i+2])
	}
	return d ^ (d >> 16)
}

