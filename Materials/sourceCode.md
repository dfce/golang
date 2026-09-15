#  Go源码

- go build -gcflags="-m" // 查看内存逃逸
- go build -gcflags="-S" // 汇编

## 
```
Go源码
│
├── 第一阶段：语言基础实现（最简单）
│      errors
│      strings
│      bytes
│      strconv
│      sort
│      sync
│      context
│
├── 第二阶段：并发源码
│      channel
│      mutex
│      rwmutex
│      waitgroup
│      cond
│      atomic
│      pool
│
├── 第三阶段：内存管理
│      new
│      make
│      slice
│      map
│      interface
│      escape
│
├── 第四阶段：runtime
│      G
│      M
│      P
│      scheduler
│      netpoll
│      timer
│
├── 第五阶段：GC
│      Root
│      Mark
│      Sweep
│      Tri-color
│      Barrier
│
├── 第六阶段：编译器
│      SSA
│      inline
│      escape
│      register
│
└── 第七阶段：标准库
       net/http
       database/sql
       tls
       crypto
       reflect
```

## slice runtime/slice.go
```go
slice 本身不存数据，只是“数据视图”
扩容机制 cap < 256 按照 *2 扩容， 超过则 按照 *1.25倍。 避免一直*2 内存浪费

append 是否会修改原slice
在未触发扩容时底层数据一样；触发扩容时则完全独立

slice header 只有 24 bytes（64位）
type slice struct {
    array unsafe.Pointer // 指向底层数组
    len int // 当前长度
    cap int // 容量
}
```

## map runtime/map.go
```go
type hmap struct {
    count       int
    flags       uint8
    B           uint8
    buckets     unsafe.Pointer
    oldBuckets  unsafe.Pointer
    nevacuate   uintptr
}
```

## channel runtime/channel.go
```go
channl = 带锁的环形队列 + 等待队列 + goroutine 阻塞/唤醒机制
type hchan struct {
    qcount      uint
    dataqsiz    unit
    buf         unsafe.Pointer

    sendx       uint
    recvx       uint

    recvq       waitq
    sendq       waitq

    lock        mutex
}

```

## mutex sync/mutex.go
- Mutex = 一个int32状态位+一个信号量（semaphore）
state 的位设计
32bit state:
| locked | woken | starving | waiters count |
|--------|-------|----------|----------------|
   1bit     1bit     1bit        29bit
- 字段含义  
locked      是否被锁
woken       是否已有 goroutine 被唤醒
starving    是否进入饥饿模式
waiters     等待队列数量 
```go
type Mutext struct {
    state int32
    sema  uint32
}
```

##  GMP 调度源码 runtime/proc.go
-   map = 数据结构核心
-   channel = 并通知核心
-   mutex =  并发控制核心
-   GMP =  Go并发能力 引擎
```go
G = Goroutine(任务)
M = Machine(线程)
P = Processor(调度器)

// G
type g struct {
    stack   stack
    sched   gobuf
    goid    int64
    status  uint32
}

// 状态
_Gidle
_Grunnable
_Gruning
_Gwationg
_Gdead

// M
type m struct {
    g0      *g
    curg    *g
    p       *g
}
// P
type p struct {
    runq    [256]g
    runnext g
}

```

##  runtime/
```go

```


##  runtime/
```go

```

##  runtime/
```go

```

##  runtime/
```go

```

##  runtime/
```go

```