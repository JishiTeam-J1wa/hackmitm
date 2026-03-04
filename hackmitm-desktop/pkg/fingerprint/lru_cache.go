package fingerprint

import (
	"container/list"
	"sync"
	"time"
)

// LRUCache LRU缓存实现
type LRUCache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
	mutex    sync.RWMutex
	ttl      time.Duration
}

// CacheEntry 缓存条目
type CacheEntry struct {
	key       string
	value     []string
	timestamp time.Time
}

// NewLRUCache 创建新的LRU缓存
func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	if capacity <= 0 {
		capacity = 1000 // 默认容量
	}
	if ttl <= 0 {
		ttl = 30 * time.Minute // 默认TTL为30分钟
	}

	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
		ttl:      ttl,
	}
}

// Get 获取缓存值
func (lru *LRUCache) Get(key string) ([]string, bool) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if elem, exists := lru.cache[key]; exists {
		entry := elem.Value.(*CacheEntry)

		// 检查是否过期
		if time.Since(entry.timestamp) > lru.ttl {
			lru.removeElement(elem)
			return nil, false
		}

		// 移动到链表头部（最近使用）
		lru.list.MoveToFront(elem)
		return entry.value, true
	}

	return nil, false
}

// Put 添加缓存值
func (lru *LRUCache) Put(key string, value []string) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if elem, exists := lru.cache[key]; exists {
		// 更新现有条目
		entry := elem.Value.(*CacheEntry)
		entry.value = value
		entry.timestamp = time.Now()
		lru.list.MoveToFront(elem)
		return
	}

	// 添加新条目
	entry := &CacheEntry{
		key:       key,
		value:     value,
		timestamp: time.Now(),
	}

	elem := lru.list.PushFront(entry)
	lru.cache[key] = elem

	// 检查容量限制
	if lru.list.Len() > lru.capacity {
		// 移除最久未使用的条目
		oldest := lru.list.Back()
		if oldest != nil {
			lru.removeElement(oldest)
		}
	}
}

// Remove 移除缓存值
func (lru *LRUCache) Remove(key string) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if elem, exists := lru.cache[key]; exists {
		lru.removeElement(elem)
	}
}

// Clear 清空缓存
func (lru *LRUCache) Clear() {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	lru.cache = make(map[string]*list.Element)
	lru.list.Init()
}

// Size 返回缓存大小
func (lru *LRUCache) Size() int {
	lru.mutex.RLock()
	defer lru.mutex.RUnlock()

	return lru.list.Len()
}

// removeElement 移除链表元素（内部方法，调用时需要已持有锁）
func (lru *LRUCache) removeElement(elem *list.Element) {
	entry := elem.Value.(*CacheEntry)
	delete(lru.cache, entry.key)
	lru.list.Remove(elem)
}

// CleanExpired 清理过期条目
func (lru *LRUCache) CleanExpired() int {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	var toRemove []*list.Element
	now := time.Now()

	// 从后往前遍历，收集过期条目
	for elem := lru.list.Back(); elem != nil; elem = elem.Prev() {
		entry := elem.Value.(*CacheEntry)
		if now.Sub(entry.timestamp) > lru.ttl {
			toRemove = append(toRemove, elem)
		} else {
			// 由于链表是按时间排序的，后面的条目都不会过期
			break
		}
	}

	// 移除过期条目
	for _, elem := range toRemove {
		lru.removeElement(elem)
	}

	return len(toRemove)
}

// GetStats 获取缓存统计信息
func (lru *LRUCache) GetStats() map[string]interface{} {
	lru.mutex.RLock()
	defer lru.mutex.RUnlock()

	return map[string]interface{}{
		"capacity":    lru.capacity,
		"size":        lru.list.Len(),
		"utilization": float64(lru.list.Len()) / float64(lru.capacity) * 100,
		"ttl_seconds": int(lru.ttl.Seconds()),
	}
}

// SetCapacity 设置缓存容量
func (lru *LRUCache) SetCapacity(capacity int) {
	if capacity <= 0 {
		return
	}

	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	lru.capacity = capacity

	// 如果新容量小于当前大小，需要移除多余的条目
	for lru.list.Len() > lru.capacity {
		oldest := lru.list.Back()
		if oldest != nil {
			lru.removeElement(oldest)
		}
	}
}

// SetTTL 设置缓存TTL
func (lru *LRUCache) SetTTL(ttl time.Duration) {
	if ttl <= 0 {
		return
	}

	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	lru.ttl = ttl
}
