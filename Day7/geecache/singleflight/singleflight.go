package singleflight

import "sync"

//调用正在进行或已完成的 Do 调用
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex
	m  map[string]*call //惰性初始化
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait() //说明其他协程有在发送当前请求，固然这个只需等待那个协程完成得响应即可
		return c.val, c.err
	}
	//等到有用到得时候再进行初始化
	c := new(call)
	//将计数器+1
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()
	//执行完函数就解锁
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
	return c.val, c.err
}
