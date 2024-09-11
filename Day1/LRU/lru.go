package LRU

import "container/list"

type Cache struct {
	maxBytes  int64 //最大字节数
	nbytes    int64
	ll        *list.List //链表
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value) //对应的一个回调函数
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

// 创建一个LRU的结构
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// 删除对应旧的结点
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele) //这里并没有实际意义上的删除，gc机制会在作用域结束之后自动帮你删除操作
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}

	}

}

// 增加值的函数（将对应内容哈希查找到，然后删除，并放到头部）
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		//这行代码是将 ele.Value 强制转换为 *entry 类型
		kv := ele.Value.(*entry) //进行断言，认为这个是entry类型，然后将其转化成entry类型
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())

	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// 对应查找
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	//value 的类型是 Value（接口），因此默认会返回 nil。
	//ok 的类型是 bool，因此默认会返回 false。
	return
}

// 实现了Cache的长度方法,对应值的接口实现在测试类里面有，可以供用户自定义长度
func (c *Cache) Len() int {
	return c.ll.Len()
}
