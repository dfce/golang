package cases

import (
	"fmt"
	"sync"
	"time"
)

/*
channel (通道) 是实现go核心并发的纽带
	“不要通过共享内存来通信，而应该通过通信来共享内存。”
	(Do not communicate by sharing memory; instead, share memory by communicating.)

Channel 必须使用make函数进行初始化。通过<-来表示数据流动

*/

func ChannelCase() {
	// buffer0Test()
	// buffer3Test()
	// ctrChanTest()
	// ctr3ChanTest()
	testInTurn3()
}

func basic() {
	ch := make(chan int) //  创建一个传递int类型的通道
	ch <- 10             // 【发送】数据10流入通道（写）
	data := <-ch         // 【接收】数据从通道流出并赋值给变量（读）
	fmt.Println(data)
	<-ch      // 【丢弃】接收数据但不使用
	close(ch) // 【关闭】关闭通道，禁止再发送数据
}

// 无缓冲vs有缓冲
// 根据初始化时是否指定容量(Capcity)
/*
	无缓冲：
ch := make(chan int) // 未指定容量， 容量=0
特点：发送和接收是同步的
阻塞规则：发送方写数据时，必须有接收方正在等待接收，否则发送方一直阻塞在当前代码行；
反之，接收方读数据时，若没有数据，也会一直阻塞。

	有缓冲
ch := make(chan int, 3) // 指定容量=3
特点：发送和接收是异步的
阻塞规则：只要通道内的元素数量未满（<3）,发送方就可以无阻塞的写，缓冲区填满时才阻塞。
缓冲区为空时，接收方才阻塞
*/
// 无缓冲：
func buffer0Test() {
	ch := make(chan string)
	go func() {
		fmt.Println("【协程】正在处理业务逻辑...")
		time.Sleep(2 * time.Second) // 模拟耗时任务
		ch <- "计算结果_SUCCESS"        // 任务完成，投递结果
		fmt.Println("【协程】结果已成功投递，并退出")
	}()
	fmt.Println("【主线程】现在去通道拿结果，如果它没写完，干等(阻塞)")
	result := <-ch
	fmt.Println("【主线程】成功拿到结果:", result)
}

// 单向通道 single
func producer(out chan<- int) {
	defer close(out)
	// chan<- 只写
	out <- 42
}
func consumer(in <-chan int) {
	// <-chan 只读
	data, ok := <-in
	if !ok {
		fmt.Println("通道关闭， 且数据已全部读完")
	}
	fmt.Println(data)
}

// 有缓冲：
// 并发频率限制器（Rate Limiter）
// 如 只能并发执行3个请求， 防止数据库被瞬间冲垮，可以利用有缓冲通道容量特性，将其作为令牌
func buffer3Test() {
	// 创建容量=3的通道，代表最大并发3
	limitCh := make(chan struct{}, 3)
	tasks := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

	for _, taskId := range tasks {
		// 往里面塞一个结构体占位，如果已满，这里就会阻塞
		// 只有等前面任务执行完成，释放了位置，后面的任务才能继续
		limitCh <- struct{}{}

		go func(id int) {
			defer func() {
				<-limitCh // 任务执行完毕，从通道继续读
			}()
			time.Sleep(30 * time.Microsecond)
			fmt.Printf("开始执行高负载并发任务：%d\n", id)
		}(taskId)
	}
}

// chan 控制多协程顺序输出奇偶数
// 3个chan
func ctr3ChanTest() {
	oddch := make(chan struct{})
	evench := make(chan struct{})
	donech := make(chan struct{})

	go func() {
		for i := 1; i <= 10; i += 2 {
			<-oddch
			fmt.Println("奇数：", i)
			evench <- struct{}{}
		}
	}()
	go func() {
		for i := 2; i <= 10; i += 2 {
			<-evench
			fmt.Println("偶数：", i)
			if i >= 10 {
				donech <- struct{}{}
			}
			oddch <- struct{}{}
		}
	}()

	oddch <- struct{}{}
	<-donech
}

// 2个chan + wait
func ctrChanTest() {
	oddch := make(chan struct{})
	evench := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(2)

	go odd(oddch, evench, &wg, 10)
	go even(oddch, evench, &wg, 10)
	oddch <- struct{}{}

	wg.Wait()
}

func odd(odd <-chan struct{}, even chan<- struct{}, wg *sync.WaitGroup, max int) {
	defer wg.Done()
	for i := 1; i <= max; i += 2 {
		<-odd
		fmt.Println("奇数:", i)
		even <- struct{}{}
	}
}
func even(odd, even chan struct{}, wg *sync.WaitGroup, max int) {
	defer wg.Done()
	for i := 2; i <= max; i += 2 {
		<-even

		fmt.Println("偶数:", i)
		if i < max {
			odd <- struct{}{}
		}
	}
}

func testInTurn3() {
	chOdd := make(chan int, 1)
	chEven := make(chan int, 1)
	doneCh := make(chan struct{})

	collect := []int{}

	go func() {
		for i := 1; i <= 10; i += 2 {
			chOdd <- i
			fmt.Println("odd:", i)
			collect = append(collect, i)
			<-chEven
		}
	}()
	go func() {
		for i := 2; i <= 10; i += 2 {
			<-chOdd
			fmt.Println("eve:", i)
			collect = append(collect, i)
			if i >= 10 {
				doneCh <- struct{}{}
			}
			chEven <- i
		}
	}()

	<-doneCh
	fmt.Println("collect", collect)
	fmt.Println("end -")
}

func testInTurn4() {
	chOdd := make(chan int, 1)
	chEven := make(chan int, 1)

	collect := []int{}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 1; i <= 10; i += 2 {
			chOdd <- i
			fmt.Println("odd:", i)
			collect = append(collect, i)
			<-chEven
		}
	}()
	go func() {
		defer wg.Done()
		for i := 2; i <= 10; i += 2 {
			<-chOdd
			fmt.Println("eve:", i)
			collect = append(collect, i)
			chEven <- i
		}
	}()

	wg.Wait()
	fmt.Println("collect", collect)

	fmt.Println("end -")
}
