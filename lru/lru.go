/*
LRU Cache
基于LRU淘汰策略的内存缓存，get/set/remove时间复杂度为O(1)
 */

package lru

import "container/list"

type Value interface {
	Len() int
}

// 链表中节点存储的数据，k用于移除element时，删除map中对应的value
type Entry struct {
	k string
	v Value
}

type Cache struct {
	maxBytes int64 // 最大Size
	cBytes int64 // 当前大小

	OnEvicted func(k string, v Value) // 淘汰回调

	ll *list.List // 双向链表，用于存储element之前的顺序
	cc map[string]*list.Element // map，存储key到element节点的映射
}

func New(maxBytes int64, OnEvicted func(k string, v Value)) *Cache {
	return &Cache{
		maxBytes: maxBytes,
		OnEvicted: OnEvicted,
	}
}

func (c *Cache)Get(key string) (v Value, ok bool) {
	if c.cc == nil {
		return
	}
	e, ok := c.cc[key]
	if ok {
		c.ll.MoveToFront(e)
		ety := e.Value.(*Entry)
		v = ety.v
		return
	}
	return
}

func (c *Cache)Remove(key string) {
	if c.cc == nil {
		return
	}
	if e, ok := c.cc[key]; ok {
		c.removeElement(e)
	}
}

func (c *Cache)Add(k string, v Value) {
	if c.cc == nil {
		c.cc = make(map[string]*list.Element)
		c.ll = list.New()
	}
	if e, ok := c.cc[k]; ok { // update
		c.ll.MoveToFront(e)
		ety := e.Value.(*Entry)
		c.cBytes += int64(v.Len() - ety.v.Len())
		ety.v = v
	} else { // insert
		e := c.ll.PushFront(&Entry{k: k, v: v})
		c.cc[k] = e
		c.cBytes += int64(len(k) + v.Len())
	}
	for c.maxBytes != 0 && c.cBytes > c.maxBytes {
		c.removeOldest()
	}
}

func (c *Cache)Len() int {
	return c.ll.Len()
}

func (c *Cache)removeOldest() {
	if c.cc == nil {
		return
	}
	b := c.ll.Back()
	if b != nil {
		c.removeElement(b)
	}
}

func (c *Cache)removeElement(e *list.Element) {
	c.ll.Remove(e)
	ety := e.Value.(*Entry)
	delete(c.cc, ety.k)
	c.cBytes -= int64(len(ety.k) + ety.v.Len())

	if c.OnEvicted != nil {
		c.OnEvicted(ety.k, ety.v)
	}
}