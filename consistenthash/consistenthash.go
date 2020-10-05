/*
实现一致性哈希算法
1.创建Hash，带虚拟倍数
2.添加节点
 */
package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash Hash // hash算法
	replicas int // 虚拟节点倍数
	keys []int // 所有的节点
	hashMap map[int]string // 虚拟节点和真实节点的映射
}

func NewMap(r int, hash Hash) *Map {
	m := &Map{
		replicas: r,
		hash: hash,
		hashMap: make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map)Add(ks...string) {
	for _, k := range ks {
		for i := 0; i<m.replicas; i++ {
			kn := strconv.Itoa(i) + k // 拼接虚拟节点key
			h := int(m.hash([]byte(kn))) // 计算虚拟节点hash值
			m.keys = append(m.keys, h) // 添加虚拟节点
			m.hashMap[h] = k // 存储虚拟节点和真实节点的映射
		}
	}
	sort.Ints(m.keys)
}

func (m *Map)Get(k string) string {

	if len(m.keys) == 0 {
		return ""
	}

	h := int(m.hash([]byte(k))) // 计算key的hash值
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= h
	}) // 搜索key应该在的虚拟节点
	if idx >= len(m.keys) {
		idx = 0 // hash环
	}
	return m.hashMap[m.keys[idx]] // 取虚拟节点对应的真是节点
}