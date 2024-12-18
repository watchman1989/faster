package faster

// Cache 结构体定义了缓存的基本属性
type Cache struct {
	capacity    int
	maxLifetime int64
	maxIdleTime int64
	items       map[string]string
}

// NewCache 创建一个新的缓存实例
func NewCache(capacity int, maxLifetime int64, maxIdleTime int64) *Cache {
	return &Cache{
		capacity:    capacity,
		maxLifetime: maxLifetime,
		maxIdleTime: maxIdleTime,
		items:       make(map[string]string),
	}
}

// Set 添加或更新缓存项
func (c *Cache) Set(key string, value string, ttl int64) {
	c.items[key] = value
}

// Get 获取缓存项
func (c *Cache) Get(key string) string {
	if value, exists := c.items[key]; exists {
		return value
	}
	return ""
}
