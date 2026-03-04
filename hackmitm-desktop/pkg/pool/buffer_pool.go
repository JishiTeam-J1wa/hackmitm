// Package pool 提供高效的内存池管理
// Package pool provides efficient memory pool management
package pool

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// BufferPool 高效的缓冲池管理器
type BufferPool struct {
	// pools 不同大小的缓冲池
	pools map[int]*sync.Pool
	// sizes 支持的缓冲区大小
	sizes []int
	// stats 统计信息
	stats *PoolStats
	// mutex 保护并发访问
	mutex sync.RWMutex
	// gcTicker GC定时器
	gcTicker *time.Ticker
	// stopGC 停止GC信号
	stopGC chan struct{}
}

// PoolStats 池统计信息
type PoolStats struct {
	// 分配统计
	TotalAllocated int64 // 总分配次数
	TotalReleased  int64 // 总释放次数
	CurrentInUse   int64 // 当前使用中的缓冲区数量

	// 大小统计
	SizeDistribution map[int]int64 // 各大小缓冲区的分配次数

	// 性能统计
	HitRate   float64 // 命中率
	TotalHits int64   // 总命中次数
	TotalMiss int64   // 总未命中次数

	// 内存统计
	TotalMemoryAllocated int64 // 总内存分配量
	TotalMemoryReleased  int64 // 总内存释放量

	// 时间统计
	LastGCTime time.Time // 最后GC时间
	GCCount    int64     // GC次数
}

// Buffer 缓冲区包装器
type Buffer struct {
	data []byte
	size int
	pool *BufferPool
}

// 默认缓冲区大小配置
var defaultSizes = []int{
	1024,    // 1KB
	4096,    // 4KB
	8192,    // 8KB
	16384,   // 16KB
	32768,   // 32KB
	65536,   // 64KB
	131072,  // 128KB
	262144,  // 256KB
	524288,  // 512KB
	1048576, // 1MB
	2097152, // 2MB
	4194304, // 4MB
}

// NewBufferPool 创建新的缓冲池
func NewBufferPool(sizes []int) *BufferPool {
	if sizes == nil {
		sizes = defaultSizes
	}

	bp := &BufferPool{
		pools:  make(map[int]*sync.Pool),
		sizes:  sizes,
		stopGC: make(chan struct{}),
		stats: &PoolStats{
			SizeDistribution: make(map[int]int64),
			LastGCTime:       time.Now(),
		},
	}

	// 初始化各大小的缓冲池
	for _, size := range sizes {
		bp.initPool(size)
	}

	// 启动GC定时器
	bp.startGC()

	return bp
}

// initPool 初始化指定大小的缓冲池
func (bp *BufferPool) initPool(size int) {
	bp.pools[size] = &sync.Pool{
		New: func() interface{} {
			atomic.AddInt64(&bp.stats.TotalMiss, 1)
			atomic.AddInt64(&bp.stats.TotalMemoryAllocated, int64(size))
			return &Buffer{
				data: make([]byte, size),
				size: size,
				pool: bp,
			}
		},
	}
}

// Get 获取指定大小的缓冲区
func (bp *BufferPool) Get(size int) *Buffer {
	// 找到合适的缓冲区大小
	actualSize := bp.findBestSize(size)

	atomic.AddInt64(&bp.stats.TotalAllocated, 1)
	atomic.AddInt64(&bp.stats.CurrentInUse, 1)

	bp.mutex.RLock()
	pool, exists := bp.pools[actualSize]
	bp.mutex.RUnlock()

	if !exists {
		// 如果没有合适的池，创建新的缓冲区
		atomic.AddInt64(&bp.stats.TotalMiss, 1)
		atomic.AddInt64(&bp.stats.TotalMemoryAllocated, int64(actualSize))
		return &Buffer{
			data: make([]byte, actualSize),
			size: actualSize,
			pool: bp,
		}
	}

	// 从池中获取缓冲区
	buffer := pool.Get().(*Buffer)
	buffer.data = buffer.data[:size] // 调整到实际需要的大小

	atomic.AddInt64(&bp.stats.TotalHits, 1)

	// 更新大小分布统计
	bp.mutex.Lock()
	bp.stats.SizeDistribution[actualSize]++
	bp.mutex.Unlock()

	// 更新命中率
	bp.updateHitRate()

	return buffer
}

// Put 释放缓冲区回池中
func (bp *BufferPool) Put(buffer *Buffer) {
	if buffer == nil || buffer.pool != bp {
		return
	}

	atomic.AddInt64(&bp.stats.TotalReleased, 1)
	atomic.AddInt64(&bp.stats.CurrentInUse, -1)
	atomic.AddInt64(&bp.stats.TotalMemoryReleased, int64(buffer.size))

	// 重置缓冲区
	buffer.data = buffer.data[:cap(buffer.data)]

	// 清零缓冲区内容（安全考虑）
	for i := range buffer.data {
		buffer.data[i] = 0
	}

	bp.mutex.RLock()
	pool, exists := bp.pools[buffer.size]
	bp.mutex.RUnlock()

	if exists {
		pool.Put(buffer)
	}
}

// GetBytes 获取字节切片（兼容原有接口）
func (bp *BufferPool) GetBytes(size int) []byte {
	buffer := bp.Get(size)
	return buffer.data
}

// PutBytes 释放字节切片（兼容原有接口）
func (bp *BufferPool) PutBytes(data []byte) {
	// 这里我们无法直接释放，因为没有Buffer包装器
	// 这个方法主要用于兼容性
}

// findBestSize 找到最合适的缓冲区大小
func (bp *BufferPool) findBestSize(size int) int {
	for _, poolSize := range bp.sizes {
		if poolSize >= size {
			return poolSize
		}
	}
	// 如果没有找到合适的大小，返回最大的
	if len(bp.sizes) > 0 {
		return bp.sizes[len(bp.sizes)-1]
	}
	return size
}

// updateHitRate 更新命中率
func (bp *BufferPool) updateHitRate() {
	totalHits := atomic.LoadInt64(&bp.stats.TotalHits)
	totalMiss := atomic.LoadInt64(&bp.stats.TotalMiss)
	total := totalHits + totalMiss

	if total > 0 {
		bp.stats.HitRate = float64(totalHits) / float64(total)
	}
}

// startGC 启动垃圾回收
func (bp *BufferPool) startGC() {
	bp.gcTicker = time.NewTicker(5 * time.Minute) // 每5分钟执行一次GC

	go func() {
		for {
			select {
			case <-bp.gcTicker.C:
				bp.runGC()
			case <-bp.stopGC:
				return
			}
		}
	}()
}

// runGC 执行垃圾回收
func (bp *BufferPool) runGC() {
	// 强制GC
	runtime.GC()

	// 更新统计信息
	bp.stats.LastGCTime = time.Now()
	atomic.AddInt64(&bp.stats.GCCount, 1)

	// 清理部分缓冲池（保留一定数量的缓冲区）
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	for _, pool := range bp.pools {
		// 清理池中的部分对象
		for i := 0; i < 10; i++ { // 最多清理10个对象
			obj := pool.Get()
			if obj == nil {
				break
			}
			// 不放回池中，让GC回收
		}
	}
}

// Stop 停止缓冲池
func (bp *BufferPool) Stop() {
	if bp.gcTicker != nil {
		bp.gcTicker.Stop()
	}
	close(bp.stopGC)
}

// GetStats 获取统计信息
func (bp *BufferPool) GetStats() *PoolStats {
	// 创建统计信息的副本
	stats := &PoolStats{
		TotalAllocated:       atomic.LoadInt64(&bp.stats.TotalAllocated),
		TotalReleased:        atomic.LoadInt64(&bp.stats.TotalReleased),
		CurrentInUse:         atomic.LoadInt64(&bp.stats.CurrentInUse),
		HitRate:              bp.stats.HitRate,
		TotalHits:            atomic.LoadInt64(&bp.stats.TotalHits),
		TotalMiss:            atomic.LoadInt64(&bp.stats.TotalMiss),
		TotalMemoryAllocated: atomic.LoadInt64(&bp.stats.TotalMemoryAllocated),
		TotalMemoryReleased:  atomic.LoadInt64(&bp.stats.TotalMemoryReleased),
		LastGCTime:           bp.stats.LastGCTime,
		GCCount:              atomic.LoadInt64(&bp.stats.GCCount),
		SizeDistribution:     make(map[int]int64),
	}

	// 复制大小分布统计
	for size, count := range bp.stats.SizeDistribution {
		stats.SizeDistribution[size] = atomic.LoadInt64(&count)
	}

	return stats
}

// Bytes 返回缓冲区的字节切片
func (b *Buffer) Bytes() []byte {
	return b.data
}

// Len 返回缓冲区长度
func (b *Buffer) Len() int {
	return len(b.data)
}

// Cap 返回缓冲区容量
func (b *Buffer) Cap() int {
	return cap(b.data)
}

// Size 返回缓冲区大小
func (b *Buffer) Size() int {
	return b.size
}

// Reset 重置缓冲区
func (b *Buffer) Reset() {
	b.data = b.data[:0]
}

// Write 写入数据到缓冲区
func (b *Buffer) Write(data []byte) (int, error) {
	b.data = append(b.data, data...)
	return len(data), nil
}

// Release 释放缓冲区
func (b *Buffer) Release() {
	if b.pool != nil {
		b.pool.Put(b)
	}
}

// 全局缓冲池实例
var (
	globalBufferPool *BufferPool
	poolOnce         sync.Once
)

// GetGlobalPool 获取全局缓冲池实例
func GetGlobalPool() *BufferPool {
	poolOnce.Do(func() {
		globalBufferPool = NewBufferPool(nil)
	})
	return globalBufferPool
}

// GetBuffer 从全局池获取缓冲区
func GetBuffer(size int) *Buffer {
	return GetGlobalPool().Get(size)
}

// PutBuffer 释放缓冲区到全局池
func PutBuffer(buffer *Buffer) {
	GetGlobalPool().Put(buffer)
}

// GetBytes 从全局池获取字节切片
func GetBytes(size int) []byte {
	return GetGlobalPool().GetBytes(size)
}

// PutBytes 释放字节切片到全局池
func PutBytes(data []byte) {
	GetGlobalPool().PutBytes(data)
}
