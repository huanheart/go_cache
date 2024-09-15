package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 定义一个函数类型
type Hash func(data []byte) uint32

type Map struct {
	hash     Hash
	replicas int            //虚拟结点要创建多少个,即一个真实结点拥有多少个虚拟结点可以映射到这个真实结点
	keys     []int          //表示一个哈希环
	hashMap  map[int]string //一个虚拟结点映射到哪一个真实结点

}

// 创建一致性哈希
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 增加一些keys到hash中
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			//创建replicas个虚拟结点hash,利用自定义哈希算法m.hash将其对应的位置返回并转化成int
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			//将其虚拟结点加入到对应的keys上
			m.keys = append(m.keys, hash)
			//将对应虚拟结点映射到真实结点
			m.hashMap[hash] = key
		}
		//将对应元素排序
		sort.Ints(m.keys)
	}
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	//将key转化成切片，然后进行返回对应的虚拟结点
	hash := int(m.hash([]byte(key)))
	//在已排序的切片 m.keys 中查找第一个大于或等于 hash 的元素的索引。
	//如果找到了这样的元素，idx 将是该元素的索引；如果没有找到，idx 将是 len(m.keys)，即切片的长度。
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	//找到对应虚拟结点映射的真实结点
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
