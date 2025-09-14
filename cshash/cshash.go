package cshash

import (
	"hash/fnv"
	"sort"
	"strconv"
	"sync"
)

type HashFunc func([]byte) uint64

type HashMap struct {
	nodes       []virtualNode
	mu          sync.RWMutex
	virtualNums int
	hashFunc    HashFunc
}

type virtualNode struct {
	virtualKey string
	key        string
	value      uint64
}

func NewMap(virtualNums int, hashFunc HashFunc) *HashMap {
	if hashFunc == nil {
		hashFunc = func(data []byte) uint64 {
			h := fnv.New64a()
			h.Write(data)
			return h.Sum64()
		}
	}
	return &HashMap{
		virtualNums: virtualNums,
		hashFunc:    hashFunc,
	}
}

func (hm *HashMap) Update(deletes []string, inserts []string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	// 标记待删除节点
	deleteSet := make(map[string]bool)
	for _, key := range deletes {
		deleteSet[key] = true
	}

	// 生成新虚拟节点
	newNodes := make([]virtualNode, 0, len(inserts)*hm.virtualNums)
	for _, key := range inserts {
		for j := 0; j < hm.virtualNums; j++ {
			virtualKey := key + "_" + strconv.Itoa(j)
			newNodes = append(newNodes, virtualNode{
				virtualKey: virtualKey,
				key:        key,
				value:      hm.hashFunc([]byte(virtualKey)),
			})
		}
	}

	// 保留未删除的旧节点
	for _, node := range hm.nodes {
		if !deleteSet[node.key] {
			newNodes = append(newNodes, node)
		}
	}

	// 排序并更新环
	sort.Slice(newNodes, func(i, j int) bool {
		return newNodes[i].value < newNodes[j].value
	})
	hm.nodes = newNodes
}

func (hm *HashMap) Get(key string) string {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if len(hm.nodes) == 0 {
		return ""
	}
	hash := hm.hashFunc([]byte(key))
	index := sort.Search(len(hm.nodes), func(i int) bool {
		return hm.nodes[i].value >= hash
	})
	return hm.nodes[index%len(hm.nodes)].key
}
