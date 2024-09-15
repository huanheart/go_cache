package geecache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc((func(key string) ([]byte, error) {
		return []byte(key), nil
	}))
	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Fatal("callback failed")
	}

}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	//GetterFunc这里的回调函数什么时候会被触发？
	//答：当使用下面的Get函数，一开始时发现没有这个函数的，所以会将其置为0，但是每一个key都调用了两次Get函数
	//固然此时会变成1最终
	gee := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			//必须要在db这个map中，这个key，否则就被置为0
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key]++
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	//查找对应的map与之前存放的map中的key以及值是否一致，如果出现了key一致，但是value不一致，那么就说明错误
	for k, v := range db {
		if view, err := gee.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		}
		//如果键对应的哈希值超过1，也说明错误了
		if _, err := gee.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}

	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}

}

func TestGetGroup(t *testing.T) {
	groupName := "scores"
	NewGroup(groupName, 2<<10, GetterFunc(
		func(key string) (bytes []byte, err error) { return }))

	if group := GetGroup(groupName); group == nil || group.name != groupName {
		t.Fatalf("group %s not exist ", groupName)
	}
	if group := GetGroup(groupName + "111"); group != nil {
		t.Fatalf("expect nil, but %s got", group.name)
	}
}
