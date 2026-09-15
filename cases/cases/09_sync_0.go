package cases

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

/*
sync
├── Mutex          // 互斥锁
├── RWMutex        // 读写锁
├── WaitGroup      // 等待协程结束
├── Once           // 只执行一次
├── Cond           // 条件变量
├── Pool           // 对象池
├── Map            // 并发安全Map
├── Locker         // 锁接口
*/
func SyncCase() {
	// waitTest()
	// mutexTest()
	rwmutexTest()
}

// waitgroup
// 等待多个 goroutine 完成
func waitTest() {
	defer timeConsuming("waitTest", time.Now())
	cumulative := 0
	// wg := sync.WaitGroup{}
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			wg.Done()
			// fmt.Printf("cumulative ++ 第: %d次, 结果： %d\n", i, cumulative)
			cumulative++
		}()
	}
	wg.Wait()
	fmt.Println("cumulative:", cumulative)
}

// Mutex 互斥锁
// 只用有一个人拿到锁，任何人都需要等待(无论读写)
func mutexTest() {
	count := 0
	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			wg.Done()
			// 如果不加锁， 那么在count++的时候可能同时执行 产生的结果不准确
			mu.Lock()
			count++
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Println("count:", count)
}

// 读写锁 【读读不互斥、读写互斥、写写互斥】
type safeCache struct {
	mu    sync.RWMutex // 读写锁
	items map[string]string
}

func NewSafeCache() *safeCache {
	return &safeCache{items: make(map[string]string)}
}

// Get 读操作：使用RLock, 允许多个协程同时读取
func (c *safeCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.items[key]
	return val, ok
}

// Set 写操作：使用Lock，排它锁，任何人不能读写
func (c *safeCache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = value
}

func rwmutexTest() {
	defer timeConsuming("rwmutexTest", time.Now())
	rw := NewSafeCache()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			wg.Done()
			k := "test_" + strconv.Itoa(i)
			rw.Set(k, strconv.Itoa(i))
		}()
	}
	wg.Wait()

	t1, ok := rw.Get("test_1")
	fmt.Println("test1", t1, ok)
	t2, ok := rw.Get("test2")
	fmt.Println("test2", t2, ok)
}

func timeConsuming(name string, start time.Time) {
	fmt.Println(name, "执行耗时L：", time.Since(start))
}
