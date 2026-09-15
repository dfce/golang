package cases

import (
	"fmt"
	"sync"
	"time"
)

func CondCase() {
	list := make([]int, 0)
	cond := sync.NewCond(&sync.Mutex{})

	go readList(&list, cond)
	go readList(&list, cond)
	go readList(&list, cond)
	time.Sleep(3 * time.Second)
	initList(&list, cond)
}

func initList(list *[]int, c *sync.Cond) {
	// 主叫方, 可以持有,也可以不持有锁
	c.L.Lock()
	defer c.L.Unlock()

	for i := 0; i < 10; i++ {
		*list = append(*list, i)
	}
	// 唤醒所有协程
	c.Broadcast()
	// // 唤醒一个或者多个
	// c.Signal()
	// c.Signal()
}

func readList(list *[]int, c *sync.Cond) {
	// 被叫方
	c.L.Lock()
	defer c.L.Unlock()
	for len(*list) == 0 {
		fmt.Println("readList wait")
		c.Wait()
	}
	fmt.Println("list 数据为:", *list)
}

/*
队列
*/
type queue struct {
	list []int
	cond *sync.Cond
}

func newQueue() *queue {
	q := &queue{
		list: []int{},
		cond: sync.NewCond(&sync.Mutex{}),
	}
	return q
}

// 写入数据
func (q *queue) Put(item int) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	q.list = append(q.list, item)
	// 当数据写入成功后, 唤醒一个协程来处理数据
	q.cond.Signal()
}

// 获取数据
func (q *queue) GetMany(n int) []int {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	for len(q.list) < n {
		q.cond.Wait()
	}

	// 取走需要的数, 更新q.list
	list := q.list[:n]
	q.list = q.list[n:]
	return list
}

func CondQueueCase() {
	q := newQueue()

	for i := 0; i < 100; i++ {
		q.Put(i)
	}
	var wg sync.WaitGroup
	// 获取
	for n := 1; n <= 10; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			list := q.GetMany(n)
			fmt.Printf("%d: %d \n", n, list)
		}(n)
	}

	wg.Wait()
}
