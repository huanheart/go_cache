package singleflight

import "testing"

func TestDo(t *testing.T) {
	var g Group
	//使用这个函数，这里并没有模拟并发场景，只是简单测试了一下
	v, err := g.Do("key", func() (interface{}, error) {
		return "bar", nil
	})
	if v != "bar" || err != nil {
		t.Errorf("Do v=%v,error=%v", v, err)
	}

}
