package faster

import (
	"fmt"
	"testing"
)

func TestFaster(t *testing.T) {
	cache := NewCache(1000000, 1000000, 1000000)
	cache.Set("key1", "value1", 1000000)
	value := cache.Get("key1")
	fmt.Println(value)
}
