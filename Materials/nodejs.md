JavaScript
EventLoop
Promise
Stream
Buffer


# Nodejs

## NodeJs 的完整链路
```
                  JavaScript(V8)
                         │
                   Call Stack
                         │
                    EventLoop
                         │
        ┌────────────────┼────────────────┐
        │                                 │
   Promise(MicroTask)             Timer/MacroTask
        │                                 │
        └────────────────┬────────────────┘
                         │
                      libuv
          (epoll + ThreadPool + Queue)
                         │
                  Linux Kernel(IO)
                         │
                  Socket / File / TCP
                         │
              Buffer（二进制内存）
                         │
                  Stream（流式处理）
                         │
                HTTP / WebSocket / NestJS
```
```
                 JavaScript(V8)
                       │
             单线程执行(Call Stack)
                       │
                EventLoop(Event Loop)
                       │
          ┌────────────┴────────────┐
          │                         │
      Promise                  Stream/Buffer
      MicroTask                 IO数据处理
          │                         │
          └────────────┬────────────┘
                       │
                  libuv(epoll)
                       │
                    Linux Kernel
```

## JavaScript 单线程
```sh
1. 为什么是单线程
DOM 不能同时修改，避免竞争（设计成单线程）
```

# EventLoop

##  调用优先级
```js
/**
 * @desc Call Stack 
 */
foo() -> bar() -> console.log()

// Stack:
console.log() -> bar() -> foo() -> main()


/**
 * @desc Task Queue 
 * @desc 不会立即执行, 先进入 Timer Queue; 
 * 等待 CallStack 为空 -> EventLoop 取任务 -> 执行
 * 所以执行时机 可能比规定时间 长
 */
setTimeout(()=>) 

/**
 * @desc MicroTask 
 * @desc 优先级 ， 执行顺序 同步 -> MicroTask -> MicroTask
 */
Promise.then()
queueMicrotask()
MutationObserver(browser)

// 示例(经典)
console.log(1)
setTimeout(()=> console.log(2))
Promise.resolve().then(()=>console.log(3))
console.log(4)
// 输出结果：1、4、3、2
```

## Node EventLoop 六个阶段
```js
Timers -> Pending Callback -> Idle -> Pool -> Check -> Close Callback

// Timers
setTimeout()
setInterval()

// Poll 重要；负责：socket、file、network、accept、read

// Check 执行 setImmediate()
// 如下执行顺序不确定，IO回调里面 setImmediate 优先
setTimeout()
setImmediate()
```

## Promise
```js
/**
 * @desc Promise
 * 三种状态： 只能变化一次
 *  Pending
 *  Fulfilled
 *  Rejected
 */ 

//内部结构
Promise
status
value
reason
onFulfilled[]
onRejected[]


/**
 * @desc Promise.all()
 * @feature 要么全部成功， 返回数组； 要么整体失败
 */ 
Promise.all([
  a(), b(), c()
])

/**
 * @desc Promise.race()
 * @feature 返回最先完成的
 */ 
Promise.race([
  a(), b(), c()
])

/**
 * @desc Promise.any()
 * @feature 只要一个成功就返回； 或者全部失败
 */ 
Promise.any([
  a(), b(), c()
])

/**
 * @desc Promise.allSettled()
 * @feature 全部执行，返回所有状态， 不会reject
 */ 
Promise.allSettled([
  a(), b(), c()
])
```

# Buffer
```js
// 类型
// JavaScript:
  String、UTF8、UTF16、Unicode
// 网络
  Binary
// 所有需要 Buffer = Binary

// Buffer 本质(连续内存)
Buffer -> Uint8Array -> C++ -> Memory

// 创建Buffer
Buffer.alloc(10)
Buffer.from('hello')
Buffer.concat()
```

# Stream
```js
/**
 * @desc 不是一次全部读取, 分块读取。 不会OOM
 *  Readable
 *  Writable
 *  Duplex 双工
 *  Transform
 */ 
// Readable & Writable
fs.createReadStream() // chunk -> chunk-> ... -> end
fs.createWriteStream() // write -> finish

// Duplex: Readable + Writable。 如：
Socket

// Transform Input -> Transform -> Output。如：
gzip -> encrypt -> decode

// Pipe ReadStream -> pipe() -> WriteStream

```