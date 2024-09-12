package geecache

import (
	"fmt"
	"log"
	"sync"
)

// Group 是 GeeCache 最核心的数据结构，负责与用户的交互，并且控制缓存值存储和获取的流程
type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

// 定义了一个回调函数的接口
// 这个回调函数主要进行一个返回数据的作用，可以将当前函数所在的作用域的东西进行一个返回
type Getter interface {
	Get(key string) ([]byte, error) //需要实现一个这个函数
}

// 将其对应回调函数重新变成另外一个新类型，在go中，这个是一个新类型
type GetterFunc func(key string) ([]byte, error)

// GetterFunc类型实现了Get接口，用于调用func(key string) ([]byte,error)的这个函数
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// 建立多个缓存结构，这样可以实现缓存多种数据类型
// 全局声明需要放到最外面
var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// 创建一个新的类型的缓存结构,传入了一个接口
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// GetGroup 返回先前使用 NewGroup 创建的命名组，或者
// 如果不存在这样的组，则返回 nil。
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// 从这个缓存组中拿取对应之前缓存过的内容
func (g *Group) Get(key string) (ByteView, error) {
	//必须含有key
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}
	//如果没有这个对应的缓存，那么就从Lru里面内部拿取（即可以理解为磁盘中拿取）
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	//调用回调函数,触发没有key缓存对应的回调函数
	//这个回调函数挺关键的
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{
		b: cloneBytes(bytes),
	}
	//填充对应的缓存
	g.populateCache(key, value)
	return value, nil
}
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
