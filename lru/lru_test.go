package lru

import (
	"fmt"
	"testing"
)

type S string

func (s S)Len() int {
	return len(s)
}

var getTests = []struct{
	name string
	keyToAdd string
	keyToGet string
	expectOk bool
} {
	{"gettest_hit", "key1", "key1", true},
	{"gettest_miss", "key1", "key2", false},
}

/*
测试缓存命中
 */
func TestGet(t *testing.T) {
	for _, gt := range getTests {
		c := New(0, nil)
		c.Add(gt.keyToAdd, S("1234"))
		v, ok := c.Get(gt.keyToGet)
		if gt.expectOk != ok {
			t.Fatalf("%s cache get = %v, want = %v", gt.name, ok, gt.expectOk)
		} else if ok && v != S("1234") {
			t.Fatalf("%s expect get %v, but get %v", gt.name, S("1234"), v)
		}
	}
}

/*
测试缓存删除
 */
func TestRemove(t *testing.T) {
	lru := New(0, nil)
	lru.Add("myKey", S("1234"))
	if val, ok := lru.Get("myKey"); !ok {
		t.Fatal("TestRemove returned no match")
	} else if val != S("1234") {
		t.Fatalf("TestRemove failed.  Expected %d, got %v", 1234, val)
	}

	lru.Remove("myKey")
	if _, ok := lru.Get("myKey"); ok {
		t.Fatal("TestRemove returned a removed entry")
	}
}

/*
1. 测试缓存LRU淘汰
2. 测试淘汰回调
 */
func TestEvict(t *testing.T) {
	evictedKeys := make([]string, 0)
	onEvictedFun := func(key string, v Value) {
		evictedKeys = append(evictedKeys, key)
	}

	lru := New(50, onEvictedFun)
	for i := 0; i < 10; i++ {
		lru.Add(fmt.Sprintf("myKey%d", i), S("1234")) // 10 bytes
	}

	if len(evictedKeys) != 5 {
		t.Fatalf("got %d evicted keys; want 5", len(evictedKeys))
	}
	if evictedKeys[0] != "myKey0" {
		t.Fatalf("got %v in first evicted key; want %s", evictedKeys[0], "myKey0")
	}
	if evictedKeys[1] != "myKey1" {
		t.Fatalf("got %v in second evicted key; want %s", evictedKeys[1], "myKey1")
	}
}