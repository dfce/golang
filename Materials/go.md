
GMP 调度、内存管理、GC、Redis、MySQL、PostgreSQL、Nginx、MQ、缓存一致性、分布式事务、微服务、Kubernetes
Go基础
GMP
Goroutine
Channel
Context
Sync
Map源码
Slice源码
GC
Gin源码
Asynq

# Golang

## Slice
```go
// Slice 不是数组。
// 数组： 定长， 不能超出指定长度
arr := [5]int{1,2,3,4,5}

// 切片 动态长度
slice := []int{}
var slice1 = make([]int, 3, 5)

// 当append发生时：
如果 len < cap , 直接写入；否则：申请新的数组-> copy旧数据-> 返回新的slice
扩容规则：(Go1.18+)
  cap < 256 , cap * 2
  cap >= 256 , cap * 1.25 左右
```

## Map
```go
// map 不是线程安全
// 可能写不完，甚至出现：fatal error: map: map[int]int{}concurrent map writes
m := map[int]int{}
go func(){
		m[1] = 1
}()
go func(){
  m[2] = 2
}()

go func(){
  m[3] = 3
}()
```

## Goroutine 协程
```go
//  非线程， 而是 User Thread
// 运行： G -> P -> M -> OS Thread
// 创建方式: 只需要 2KB Stack, 线程则 1MB左右
go func (){

}() 
```

## GMP
1. G Goroutine，保存 Stack、PC、Context
2. M Machine， OS Thread 真正执行 CPU
3. P Processor，Local Queue Scheduler 
- 关系图
          G
          G
          G
          │
          ▼
+----------------+
       P
 Local Queue
+----------------+
          │
          ▼

+----------------+
       M
OS Thread
+----------------+
          │
          ▼
CPU
```
Work Stealing
如果：
P1 100G
P2 0G
P2 会分走 P1 一半的任务， 提高 CPU 利用率

```

## Channel

## Select
```go
// select 用来监听 多个 channel 的通信操作，哪个 channel准备好了就执行哪个

select {
  case <-ch1:
    fmt.Println("ch1")
  case <-ch2:
    fmt.Println("ch2")
  default:
    fmt.Println("default")
}
```

## Context
- 
  1. Context 用于取消、超时、请求级数据传递，不用于依赖注入。
  2. Context 永远作为函数第一个参数，不保存到结构体中。
  3. WithCancel、WithTimeout、WithDeadline 创建的 Context，拿到 cancel 后原则上都应调用（通常 defer cancel()）。
  4. 所有可能阻塞的 Goroutine、Channel、HTTP、数据库、RPC 调用，都应该有响应 ctx.Done() 的能力。
  5. 生产环境几乎所有长生命周期 Goroutine 都应该具备 ctx.Done() 退出路径，这是避 免 Goroutine 泄漏的标准写法。
```go
HTTP:
  Request -> Service -> DAO -> Redis -> MySQL

Context树
  Background -> WithCancel -> WithTimeout -> WithValue -> cancel()
  -> 所有child取消
典型实用

ctx, cancel := context.WithTimeout(
  context.Background(),
  5*time.Second(),
  defer cancel()
)


示例1 ：Background
ctx := context.Background()

示例2 ：WithCancel
ctx, cancel := context.WithCancel(context.Background())
go func(){
  for {
    select{
    case <- ctx.Done()
      fmt.Println("worker 退出")
      return
    default:
      fmt.Println("working")
      time.Sleep(time.Second)   
    }
  }
}()

time.Sleep(3*time.Second)
cancel()

或者 一个 context 控制多个 Goroutine
ctx, cancel := context.WithCancel(contex.BackGround())
for i := 0; i < 5; i++ {
  go worker(ctx)
}
cancel()

示例3 ：WithTimeout (调用三方接口超时)
ctx, cancel := context.WithTimeout(
  context.BackGround(),
  3*time.Second
)
defer cancel()

select {
case <- ctx.Done()
  fmt.Println(ctx.Err())  
}

示例4 ：WithContext HTTP 最佳实践
func LoginController(c *gin.Context) {
  ctx := c.Request.Context
  user, err := service.LoginService(ctx)
}
// service Login
func LoginService(ctx context.Context) {
  db.WithContext(ctx)
}

示例5 ：WithValue 
用于: requestId traceId userId language tenantId 等生命周期短的
type TraceID string
ctx := context.WithValue(
  context.BackGround(),
  TraceID("trace"),
  "uuid-str"
)
fmt.Println(ctx.Value(TraceID("trace"))) // 输出： uuid-str


```

## Sync
sync
├── Mutex（互斥锁）
├── RWMutex（读写锁）
├── WaitGroup（等待组）
├── Once（单例）
├── Cond（条件变量）
├── Map（并发Map）
├── Pool（对象池）
├── Atomic（sync/atomic）
└── Locker接口


### Mutex 防止多个 Goroutine同时修改数据， 可以用 atomic 替换实现
  - Lock()
  - Unlock()
```go
错误示例：
func main() {
  var cnt int
  for i :=0; i < 1000; i++ {
    go func(){
      cnt++
    }()
  }
  fmt.Println("cnt", cnt) // 可能是1000 或者< 1000的任意树
  // cnt++  实际上 Load -> Add -> Store 不是原子操作
} 

正确示例：
func main(){
  var (
    cnt int
    mu sync.Mutex
  )
  for i :=0; i < 1000; i++ {
    go func(){
      mu.Lock()
      defer mu.Unlock()
      cnt++
    }()
  }
}
```

### RWMutex 读多写少（10000读、10写）
```go
var lock sync.RWMutex
func Get() {
  lock.Rlock()
  defer lock.RUnlock()
}
func Set() {
  lock.Lock()
  defer lock.Unlock()
}
```
### WaitGroup
```go
var wg sync.WaitGroup
wg.Add(2)
go func(){
  defer wg.Done()
}()
go func(){
  defer wg.Done()
}()

go func(){
  wg.Wait()
}()
```

### atomic 
- 利用 CPU 提供的原子指令（CAS、XADD等），保证操作不可分割 【不是锁，而是CPU指令】
Atomic 是基于 CPU 原子指令（CAS）实现的无锁并发机制，适合单个变量、状态位、计数器、指针等高频读写场景。对于多字段一致性、复杂数据结构（map、slice、tree）和事务性更新，应优先使用 Mutex/RWMutex。生产环境中，Atomic.Pointer 配合 Copy-On-Write 是配置中心、路由表、黑白名单等高性能读多写少场景的经典方案。
 CPU -> LOCK XADD -> 一次完成 -> 不会被其他CPU打断
  go1.19后推荐使用 Typed Atomic
  - atomic.Bool
  - atomic.Int32
  - atomic.Int64
  - atomic.Uint64
  - atomic.Pointer[T]
  - atomic.Value
```go
示例 1: 计数器
import (
  "fmt"
  "sync"
  "sync/atomic"
)

func main() {
  var count atomic.Int64
  var wg async.WaitGroup
  for i := 0; i < 1000; i++ {
    wg.Add(1)
    go func() {
      defer wg.Done()
      count.Add(1)
    }()
  }
  wg.Wait()
}

示例 2: Bool 程序状态
var running atomic.Bool
running.Store(true)

for running.Load() {
  process()
}
停止时直接用 
running.Store(false)


示例 3: Pointer 更新配置
如： OldConfig -> 10000 Reader. 需要更新时 不直接修改OldConfig；
  创建新的配置 用 atomic.Pointer 一次性替换， 保障读取到都是一致的

type Config struct {
  Id int46
  Name string
}
var config atomic.Pointer[Config]
func main() {

  // 设置 读取
  config.Store(&Config{
    Id: 100,
    Name: "dev"
  })

  cfg := config.Load()
  fmt.Println(cfg.Name)

  // 更新
  config.Store(&Config{
    Id: 101,
    Name: "prod"
  })
}

示例 4: Value 适合保存任意对象
var value atomic.Value
value.Store("hello")

fmt.Println(value.Load())

```

###
```go
RWMutex
  Reader -> Reader -> Reader -> ...  -> Writer(独占)
Once
 sync.Once // 保证只执行一次  
    
WaitGroup
  Add()
  Done()
  Wait()

sync.Map (适合读多写少)

```

## GC
- 自动回收程序中不再使用的内存， GC 只管理Heap
                Process
                    │
        ┌───────────┴───────────┐
        │                       │
      Stack                  Heap
        │                       │
    Goroutine             Runtime管理
                                │
        ┌───────────────┬──────────────┐
        │               │              │
      mcache         mcentral        mheap
```go
 malloc -> 没人释放 -> OOM

# 三色标记
White   没访问， 可能垃圾
Gray    自己扫了， 孩子没扫
Black   自己和孩子都扫了

GC 完整流程
GC Start
    ↓
STW（几十微秒）
    ↓
Root Scan
    ↓
Background Mark Worker
    ↓
Mutator Assist
    ↓
Mark
    ↓
Mark Done
    ↓
STW
    ↓
Sweep
    ↓
Resume
```