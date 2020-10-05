/*
singleFlight，防止缓存击穿
 */
package singleflight

import (
	"log"
	"sync"
)

type call struct {
	wg sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex
	m map[string]*call
}

func (g *Group)Do(k string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[k]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		log.Printf("Do key %s, hit loaderGroup", k)
		return c.val, c.err
	}
	c := &call{}
	g.m[k] = c
	g.mu.Unlock()

	c.wg.Add(1)
	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, k)
	g.mu.Unlock()
	return c.val, c.err
}