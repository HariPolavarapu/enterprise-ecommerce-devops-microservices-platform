package service

type CacheService struct {
	store map[string]string
}

func NewCacheService() *CacheService {
	return &CacheService{store: make(map[string]string)}
}

func (c *CacheService) Get(key string) (string, bool) {
	value, ok := c.store[key]
	return value, ok
}

func (c *CacheService) Set(key, value string) {
	c.store[key] = value
}

func (c *CacheService) Delete(key string) {
	delete(c.store, key)
}
