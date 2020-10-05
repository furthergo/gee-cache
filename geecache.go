package geecache

import (
	"fmt"
	"github.com/futhergo/gee-cache/lru"
	"github.com/futhergo/gee-cache/singleflight"
	"log"
	"sync"
	"time"
)

// 定义Getter interface，用于设置缓存miss时获取数据的Get回调方法
type Getter interface {
	Get(k string) ([]byte, error)
}

// 定义一个和Get接口函数一样的函数别名，用来实现Getter
// 好处是可以定义匿名函数来实现Getter
type GetterFunc func(k string) ([]byte, error)

// 调用自己，实现Getter
func (f GetterFunc)Get(k string) ([]byte, error) {
	return f(k)
}

// 对应一组KV，每个Group有自己的分布式缓存peer
type Group struct {
	name string // namespace
	c cache // 本地缓存
	gt Getter // 缓存miss
	peerPicker PeerPicker // 分布式节点选择
	loadGroup flightGroup // 缓存miss时，load操作限制，防止缓存击穿：cache miss时，多个请求同时打到DB，造成DB压力瞬间增大
}

type flightGroup interface {
	Do(string, func()(interface{}, error)) (interface{}, error)
}

var (
	mu sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, gt Getter) *Group {
	if gt == nil {
		log.Fatalf("create group of name: %s with nil getter", name)
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name: name,
		gt: gt,
		c: cache{
			cacheBytes: cacheBytes,
		},
		loadGroup: &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// 获取Group
func GetGroup(name string) (g *Group, ok bool) {
	mu.RLock()
	defer mu.RUnlock()
	g, ok = groups[name]
	return
}

// 注册节点选择器
func (g *Group)RegisterPeer(peerPicker PeerPicker) {
	if g.peerPicker != nil {
		panic("register peer multiple times")
	}
	g.peerPicker = peerPicker
}

// 从Group中取Key，暴露给http服务使用
func (g *Group)Get(k string) (ByteView, error) {
	if k == "" {
		return ByteView{}, fmt.Errorf("key is nil")
	}
	if v, ok := g.c.get(k); ok { // 尝试从缓存中取
		log.Printf("[GeeCache]: hit: %s\n", k)
		return v, nil
	}
	return g.load(k) // load
}

func (g *Group)load(k string) (ByteView, error) {
	bv, err := g.loadGroup.Do(k, func() (interface{}, error) { // load kv
		time.Sleep(1*time.Second)
		if g.peerPicker == nil { // 是否有分布式选项
			return g.getLocally(k)
		}
		peer, err := g.peerPicker.PickPeer(k) // 是否能取到远端节点
		if err != nil {
			return g.getLocally(k)
		}

		bs, err := peer.Get(g.name, k) // 用PeerGetter获取对应value
		if err != nil {
			return g.getLocally(k)
		}
		return ByteView{b: bs}, err
	})

	if err != nil {
		return ByteView{}, err
	}

	return bv.(ByteView), nil
}

// 调用Getter获取本地Value并更新Cache
func (g *Group)getLocally(k string) (ByteView, error) {
	bytes, err := g.gt.Get(k)
	if err != nil {
		return ByteView{}, err
	}
	v := ByteView{
		b: cloneBytes(bytes),
	}
	g.populateCache(k, v)
	return v, nil
}

// 更新Cache
func (g *Group)populateCache(k string, v ByteView) {
	g.c.add(k, v)
}

// 并发安全cache
type cache struct {
	mu sync.RWMutex
	lru *lru.Cache
	cacheBytes int64
}

func (c *cache)add(k string, v ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(k, v)
}

func (c *cache)get(k string) (bv ByteView, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.lru == nil {
		return
	}
	v, ok := c.lru.Get(k)
	if !ok {
		return
	}
	bv = v.(ByteView)
	return
}

func (c *cache)bytes() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cacheBytes
}