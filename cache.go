package sds

import (
	"container/list"
	"time"
)

const (
	defaultExpire = time.Duration(3*3600) * time.Second

	modeLru = 0

	typeKv   = 0
	typeHash = 1
)

type entry struct {
	expiration int64
	key        string
	//data type, 0: value, 1: hash map
	dataType int
	//key value type
	value interface{}
	//hash map type
	hashMap map[string]interface{}
}

type EvictFunc func(key interface{}, value interface{})

type fasterCache struct {
	//0: LRU, 1: FIFO
	mode int
	//data map size
	size      int
	evictList *list.List
	//data store map
	dataMap map[string]*list.Element
	onEvict EvictFunc
}

func NewFasterCache(mode int, size int, onEvict EvictFunc) *fasterCache {
	if size <= 0 {
		return nil
	}
	fc := &fasterCache{
		mode:      mode,
		size:      size,
		evictList: list.New(),
		dataMap:   make(map[string]*list.Element),
		onEvict:   onEvict,
	}
	return fc
}

// clean
func (fc *fasterCache) clean() {
	for k, v := range fc.dataMap {
		delete(fc.dataMap, k)
		if fc.onEvict != nil {
			fc.onEvict(k, v)
		}
	}
	fc.evictList.Init()
}

//key value

// set key value
func (fc *fasterCache) set(key string, value interface{}, expiration time.Duration) {
	if key == "" {
		return
	}
	//key是否存在
	if ee, ok := fc.dataMap[key]; ok {
		ent := ee.Value.(*entry)
		//key 存在，数据类型不为0，移除旧数据
		if ent.dataType != typeKv {
			fc.removeElement(ee)
			//如果没有设置过期时间，使用默认过期时间
			if expiration <= 0 {
				expiration = defaultExpire
			}
			nent := &entry{
				dataType:   typeKv,
				expiration: time.Now().Add(expiration).UnixNano(),
				key:        key,
				value:      value,
			}
			e := fc.evictList.PushFront(nent)
			fc.dataMap[key] = e
		} else {
			ent.value = value
			//如果没有设置过期时间，使用默认过期时间
			if expiration <= 0 {
				expiration = defaultExpire
			}
			//更新过期时间
			ent.expiration = time.Now().Add(expiration).UnixNano()
			if fc.mode == modeLru {
				fc.evictList.MoveToFront(ee)
			}
		}
	} else {
		//如果没有设置过期时间，使用默认过期时间
		if expiration <= 0 {
			expiration = defaultExpire
		}
		ent := &entry{
			dataType:   typeKv,
			expiration: time.Now().Add(expiration).UnixNano(),
			key:        key,
			value:      value,
		}
		e := fc.evictList.PushFront(ent)
		fc.dataMap[key] = e
		if fc.evictList.Len() > fc.size {
			fc.removeTail()
		}
	}
	return
}

// get key -> value
func (fc *fasterCache) get(key string) interface{} {
	//判断key是否存在
	if e, ok := fc.dataMap[key]; ok {
		ent := e.Value.(*entry)
		//非 key value类型，返回nil
		if ent.dataType != typeKv {
			return nil
		}
		//判断key是否过期
		if ent.expiration >= time.Now().UnixNano() {
			//如果是lru模式
			if fc.mode == modeLru {
				fc.evictList.MoveToFront(e)
			}
			return ent.value
		} else {
			//如果过期，删除key
			fc.removeElement(e)
			return nil
		}
	}
	return nil
}

// exist key
func (fc *fasterCache) exist(key string) bool {
	//判断key是否存在
	if e, ok := fc.dataMap[key]; ok {
		ent := e.Value.(*entry)
		//判断key是否过期
		if ent.expiration >= time.Now().UnixNano() {
			//如果是lru模式
			if fc.mode == modeLru {
				fc.evictList.MoveToFront(e)
			}
			return true
		} else {
			//如果过期，删除key
			fc.removeElement(e)
			return false
		}
	} else {
		return false
	}
}

// dataType key
func (fc *fasterCache) dataType(key string) int {
	if e, ok := fc.dataMap[key]; ok {
		ent := e.Value.(*entry)
		return ent.dataType
	}
	return -1
}

// delete key
func (fc *fasterCache) del(key string) {
	//判断key是否存在
	if e, ok := fc.dataMap[key]; ok {
		fc.removeElement(e)
	}
}

// len key
func (fc *fasterCache) len() int {
	return len(fc.keys())
}

// keys
func (fc *fasterCache) keys() []string {
	keys := make([]string, 0)
	for k, e := range fc.dataMap {
		ent := e.Value.(*entry)
		//判断key是否过期
		if ent.expiration >= time.Now().UnixNano() {
			keys = append(keys, k)
		} else {
			//如果过期，删除key
			fc.removeElement(e)
		}
	}
	return keys
}

// incrby key
func (fc *fasterCache) incrBy(key string, incr int64) int64 {
	var (
		value int64
	)
	if e, ok := fc.dataMap[key]; ok {
		ent := e.Value.(*entry)
		if ent.dataType == typeKv {
			if ori, ook := ent.value.(int64); ook {
				value = ori + incr
				ent.value = value
				//放入队列前面
				if fc.mode == modeLru {
					fc.evictList.MoveToFront(e)
				}
				return value
			}
		}
	}
	return value
}

// get ttl
func (fc *fasterCache) getTTL(key string) int64 {
	//判断key是否存在
	if e, ok := fc.dataMap[key]; ok {
		ent := e.Value.(*entry)
		//判断key是否过期
		nowAt := time.Now().UnixNano()
		if ent.expiration >= nowAt {
			//如果没有过期
			//放入队列前面
			if fc.mode == modeLru {
				fc.evictList.MoveToFront(e)
			}
			return nowAt - ent.expiration
		} else {
			//如果过期，删除key
			fc.removeElement(e)
			return 0
		}
	}
	return 0
}

// set ttl
func (fc *fasterCache) expire(key string, expiration time.Duration) {
	if expiration < 0 {
		return
	}
	//判断key是否存在
	if e, ok := fc.dataMap[key]; ok {
		ent := e.Value.(*entry)
		//判断key是否过期
		nowAt := time.Now().UnixNano()
		if ent.expiration >= nowAt {
			ent.expiration = time.Now().Add(expiration).UnixNano()
			if fc.mode == modeLru {
				fc.evictList.MoveToFront(e)
			}
		} else {
			//如果过期，删除key
			fc.removeElement(e)
			return
		}
	}
	return
}

//hash set

// hset key subkey value
func (fc *fasterCache) hSet(key, subKey string, value interface{}, expiration time.Duration) {
	if key == "" || subKey == "" {
		return
	}
	//判断key是否存在
	if ee, ok := fc.dataMap[key]; ok {
		ent := ee.Value.(*entry)
		if ent.dataType != typeHash {
			//删除当前key value
			fc.removeElement(ee)
			//如果没有设置过期时间，使用默认过期时间
			if expiration <= 0 {
				expiration = defaultExpire
			}
			nent := &entry{
				dataType:   typeHash,
				expiration: time.Now().Add(expiration).UnixNano(),
				key:        key,
				hashMap:    make(map[string]interface{}),
			}
			nent.hashMap[subKey] = value
			e := fc.evictList.PushFront(nent)
			fc.dataMap[key] = e

		} else {
			//如果没有过期
			if ent.expiration >= time.Now().UnixNano() {
				ent.hashMap[subKey] = value
				if expiration > 0 {
					//如果设置了新的过期时间，更新过期时间
					ent.expiration = time.Now().Add(expiration).UnixNano()
				}
				fc.evictList.MoveToFront(ee)
			} else {
				//如果过期，删除key
				fc.removeElement(ee)
				//如果没有设置过期时间，使用默认过期时间
				if expiration <= 0 {
					expiration = defaultExpire
				}
				nent := &entry{
					dataType:   typeHash,
					expiration: time.Now().Add(expiration).UnixNano(),
					key:        key,
					hashMap:    make(map[string]interface{}),
				}
				nent.hashMap[subKey] = value
				e := fc.evictList.PushFront(nent)
				fc.dataMap[key] = e
			}
		}
	} else {
		//如果没有设置过期时间，使用默认过期时间
		if expiration <= 0 {
			expiration = defaultExpire
		}
		ent := &entry{
			dataType:   typeHash,
			expiration: time.Now().Add(expiration).UnixNano(),
			key:        key,
			hashMap:    make(map[string]interface{}),
		}
		ent.hashMap[subKey] = value
		e := fc.evictList.PushFront(ent)
		fc.dataMap[key] = e
		if fc.evictList.Len() > fc.size {
			fc.removeTail()
		}
	}
	return
}

// hget key, subkey
func (fc *fasterCache) hGet(key, subKey string) interface{} {
	//判断key是否存在
	if e, ok := fc.dataMap[key]; ok {
		ent := e.Value.(*entry)
		if ent.dataType == typeHash {
			//判断key是否过期
			if ent.expiration >= time.Now().UnixNano() {
				//如果没有过期，判断subKey是否存在
				if val, ook := ent.hashMap[subKey]; ook {
					if fc.mode == modeLru {
						fc.evictList.MoveToFront(e)
					}
					return val
				} else {
					if fc.mode == modeLru {
						fc.evictList.MoveToFront(e)
					}
					return nil
				}
			} else {
				//如果过期，删除key
				fc.removeElement(e)
				return nil
			}
		}

	}
	return nil
}

// hexist key subkey
func (fc *fasterCache) hExist(key, subKey string) bool {
	//判断key是否存在
	if e, ok := fc.dataMap[key]; ok {
		ent := e.Value.(*entry)
		if ent.dataType == typeHash {
			//判断key是否过期
			if ent.expiration >= time.Now().UnixNano() {
				//如果没有过期，判断subKey是否存在
				_, ook := ent.hashMap[subKey]
				if fc.mode == modeLru {
					fc.evictList.MoveToFront(e)
				}
				return ook
			} else {
				//如果过期，删除key
				fc.removeElement(e)
			}
		}
	}
	return false
}

// hdel key, subkey
func (fc *fasterCache) hDel(key, subKey string) {
	//判断key是否存在
	if e, ok := fc.dataMap[key]; ok {
		ent := e.Value.(*entry)
		if ent.dataType == typeHash {
			//判断key是否过期
			if ent.expiration >= time.Now().UnixNano() {
				//如果没有过期，判断subKey是否存在
				if _, ook := ent.hashMap[subKey]; ook {
					delete(ent.hashMap, subKey)
					//如果hMap为空，删除key
					if len(ent.hashMap) == 0 {
						fc.removeElement(e)
					} else {
						//放入队列前面
						if fc.mode == modeLru {
							fc.evictList.MoveToFront(e)
						}
					}
				}
			} else {
				//如果过期，删除key
				fc.removeElement(e)
			}
		}

	}
}

// hgetall key
func (fc *fasterCache) hGetAll(key string) map[string]interface{} {
	//判断key是否存在
	if e, ok := fc.dataMap[key]; ok {
		ent := e.Value.(*entry)
		if ent.dataType == typeHash {
			//判断key是否过期
			if ent.expiration >= time.Now().UnixNano() {
				//如果没有过期，判断subKey是否存在
				if fc.mode == modeLru {
					fc.evictList.MoveToFront(e)
				}
				return ent.hashMap
			} else {
				//如果过期，删除key
				fc.removeElement(e)
				return nil
			}
		}
	}
	return nil
}

// hlen key
func (fc *fasterCache) hLen(key string) int {
	//判断key是否存在
	if e, ok := fc.dataMap[key]; ok {
		ent := e.Value.(*entry)
		if ent.dataType == typeHash {
			//判断key是否过期
			if ent.expiration >= time.Now().UnixNano() {
				//如果没有过期
				//放入队列前面
				if fc.mode == modeLru {
					fc.evictList.MoveToFront(e)
				}
				return len(ent.hashMap)
			} else {
				//如果过期，删除key
				fc.removeElement(e)
				return 0
			}
		}
	}
	return 0
}

// hkeys key
func (fc *fasterCache) hKeys(key string) []string {
	var (
		subKeys []string
	)
	subKeys = make([]string, 0)
	//判断key是否存在
	if e, ok := fc.dataMap[key]; ok {
		ent := e.Value.(*entry)
		if ent.dataType == typeHash {
			//判断key是否过期
			if ent.expiration >= time.Now().UnixNano() {
				//如果没有过期
				for ekey, _ := range ent.hashMap {
					subKeys = append(subKeys, ekey)
				}
				//放入队列前面
				if fc.mode == modeLru {
					fc.evictList.MoveToFront(e)
				}
				return subKeys
			} else {
				//如果过期，删除key
				fc.removeElement(e)
				return subKeys
			}
		}
	}
	return subKeys
}

// hincrby key subkey
func (fc *fasterCache) hIncrBy(key, subKey string, incr int64, expiration time.Duration) (value int64) {
	if key == "" || subKey == "" {
		return
	}
	//判断key是否存在
	if e, ok := fc.dataMap[key]; ok {
		ent := e.Value.(*entry)
		if ent.dataType != typeHash {
			//fmt.Printf("key %s is not hash type\n", key)
			//删除当前key value
			fc.removeElement(e)
			//如果没有设置过期时间，使用默认过期时间
			if expiration <= 0 {
				expiration = defaultExpire
			}
			nent := &entry{
				dataType:   typeHash,
				expiration: time.Now().Add(expiration).UnixNano(),
				key:        key,
				hashMap:    make(map[string]interface{}),
			}
			nent.hashMap[subKey] = incr
			e = fc.evictList.PushFront(nent)
			fc.dataMap[key] = e
			return incr
		} else {
			//fmt.Printf("key %s is hash type\n", key)
			//如果没有过期
			if ent.expiration >= time.Now().UnixNano() {
				//fmt.Printf("key %s is not expired\n", key)
				//如果subkey存在
				if v, ook := ent.hashMap[subKey]; ook {
					if ori, oook := v.(int64); oook {
						value = ori + incr
						ent.hashMap[subKey] = value
						//放入队列前面
						if fc.mode == modeLru {
							fc.evictList.MoveToFront(e)
						}
						return value
					}
				} else {
					ent.hashMap[subKey] = incr
					//放入队列前面
					if fc.mode == modeLru {
						fc.evictList.MoveToFront(e)
					}
					return incr
				}
			} else {
				//fmt.Printf("key %s is expired\n", key)
				//如果过期，删除key
				fc.removeElement(e)
				//如果没有设置过期时间，使用默认过期时间
				if expiration <= 0 {
					expiration = defaultExpire
				}
				nent := &entry{
					dataType:   typeHash,
					expiration: time.Now().Add(expiration).UnixNano(),
					key:        key,
					hashMap:    make(map[string]interface{}),
				}
				nent.hashMap[subKey] = incr
				e = fc.evictList.PushFront(nent)
				fc.dataMap[key] = e
				return incr
			}
		}
	} else {
		//fmt.Printf("key %s is not exist\n", key)
		//如果没有设置过期时间，使用默认过期时间
		if expiration <= 0 {
			expiration = defaultExpire
		}
		ent := &entry{
			dataType:   typeHash,
			expiration: time.Now().Add(expiration).UnixNano(),
			key:        key,
			hashMap:    make(map[string]interface{}),
		}
		ent.hashMap[subKey] = incr
		e = fc.evictList.PushFront(ent)
		fc.dataMap[key] = e
		if fc.evictList.Len() > fc.size {
			fc.removeTail()
		}
		return incr
	}

	return value
}

// remove tail
func (fc *fasterCache) removeTail() {
	e := fc.evictList.Back()
	if e != nil {
		fc.removeElement(e)
	}
}

// remove element
func (fc *fasterCache) removeElement(e *list.Element) {
	fc.evictList.Remove(e)
	ent := e.Value.(*entry)
	delete(fc.dataMap, ent.key)
	if fc.onEvict != nil {
		fc.onEvict(ent.key, ent.value)
	}
}
