package LRU

import (
	"reflect"
	"testing"
)

// 使用 go test 命令可以自动运行所有测试，方便进行批量测试。你可以轻松地检查所有测试的通过与失败，而不需要手动检查输出。
// 当测试失败时，testing 包会提供详细的信息，帮助你快速定位问题。例如，t.Fatalf 会输出失败的原因和堆栈跟踪信息，而简单的输出语句不容易提供这样的上下文。
// testing 包可以与工具结合，生成测试覆盖率报告，帮助你了解代码的哪些部分被测试覆盖，哪些部分没有。
// testing 包支持并发测试和基准测试（benchmarking），这对于性能评估和优化非常重要。
type String string

func (d String) Len() int {
	return len(d)
}

func TestGet(t *testing.T) {
	lru := New(int64(100), nil)
	lru.Add("key1", String("1234"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatal("cache hit key1 =1234 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatal("cache miss key2 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	//回调函数加入的时候只会加入对应的key值
	callback := func(key string, value Value) {
		//tempKeys := append([]string(nil), keys...) // 创建一个副本
		//tempKeys = append(tempKeys, key)
		keys = append(keys, key) //传了keys一个引用过去,如果需要传一个副本，这边可以自己创建对应的临时切片,如上面的示例
	}
	lru := New(int64(10), callback)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("k2"))
	lru.Add("k3", String("k3"))
	lru.Add("k4", String("k4"))

	expect := []string{"key1", "k2"}
	if !reflect.DeepEqual(expect, keys) { //比较两个切片是否相等
		t.Fatalf("Call OnEvicted failed ,expect keys equals to %s ", expect)
	}

}

func TestAdd(t *testing.T) {
	lru := New(int64(100), nil)
	lru.Add("key", String("1"))
	lru.Add("key", String("111"))
	if lru.nbytes != int64(len("key")+len("111")) {
		t.Fatal("expected 6 but got ", lru.nbytes)
	}

}

func TestRemoveoldset(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	cap := len(k1 + k2 + v1 + v2)
	lru := New(int64(cap), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))
	if _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		t.Fatal("Remove key1 failed")
	}

}
