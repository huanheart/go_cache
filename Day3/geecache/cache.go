package geecache

import (
	"awesomeProject2/Day3/geecache/LRU"
	"sync"
)

//这个类主要是可以增加缓存以及获取缓存
//这里封装了对应的LRU这个数据结构，给他多封装一层锁,变成线程安全的缓存数据结构

type cache struct {
	mu         sync.Mutex
	lru        *LRU.Cache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = LRU.New(c.cacheBytes, nil) //new一个对应的缓存，应该有很多个吧？
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
