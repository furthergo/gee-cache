package geecache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestGetFunc(t *testing.T) {
	var f Getter = GetterFunc(func(k string) ([]byte, error) {
		return []byte(k), nil
	})

	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Fatalf("test get func error of key")
	}
}

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGet(t *testing.T) {
	lc := make(map[string]int, len(db))
	gee := NewGroup("testGee", 7, GetterFunc(func(k string) ([]byte, error) {
		log.Printf("[Getter Call]: search key: %s\n", k)
		if v, ok := db[k]; ok {
			c := 1
			if preC, ok := lc[k]; ok {
				c += preC
			}
			lc[k] = c
			return []byte(v), nil
		} else {
			return nil, fmt.Errorf("error get key: %s\n", k)
		}
	}))

	for k, v := range db {
		if bv, err := gee.Get(k); err != nil || bv.String() != v {
			t.Fatalf("Gee get key %s failed", k)
		}
		if _, err := gee.Get(k); err != nil || lc[k] > 1 {
			t.Fatalf("Gee get key %s miss cache", k)
		}
	}

	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}