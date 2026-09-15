# redis/v9 常用方法

## 
```go
red = redis.Client{}

// 写入缓存，带过期时间
err := rdb.Set(ctx, key, val, expiration).Err()
if err != nil {
  // 写入异常
}

// 读取缓存， 
// rdb.Get(ctx, key) 返回： *redis.StringCmd
val, err := rdb.Get(ctx, key).Result()
if err == redis.Nil {
  // 说明key不存在（缓存位命中/Token过期），属于正常业务逻辑
  return "", nil
}else if err != nil {
  // 真正的Redis网络故障或错误
  return "", err
}

// Incr 计数器原子递增（高并发防超卖/限流）
newNum, err := rdb.Incr(ctx, "page:viw:count").Result()

// Hash 哈希表
// 方法： rdb.HSet(ctx, key, values...)
// 可以直接平铺传入字段和值， 也可以传入一个 map[string]any
mapVal := map[string]any{
  "name": "Alex",
  "age": 18,
  "role": "admin",
}
err := rdb.HSet(ctx, "profile:10086", mapVal).Err()

err := rdb.HSet(ctx, "profile:10086", "name", "Alex", "age", 18, "role", "admin").Err()
// 读取Hash中的某一字段
// rdb.HGet(ctx, key)
userMap, err := rdb.HGet(ctx, "profile:10086").Result()
if err != nil {
  return err
}
fmt.Println("姓名：", userMap["name"])


// List 列表-双向队列,常用于消息队列、最新动态列表
// rdb.RPush(ctx, key, values...), 返回当前列表总长度(int64)
length, err := rdb.RPush(ctx, "task:queue", "task_id_001", "task_id_002").Result()
// 从左出队, .Result() 返回元素值和error. 空队列返回 redis.Nil
// rdb.LPop(ctx, key)
task, err := rdb.LPop(ctx, "task:queue").Result()
if err == redis.Nil {
  // 队列空了
}
// 阻塞式弹出（异步任务首选）
/* 
  .Result() 返回一个切片([]string),其中第一个元素 res[0]是Key的名字，第二个元素res[2]才是
  真实数据。如果超过了指定的 timeout 时间队列依然没数据， 返回 redis.Nil
  0 代表无限阻塞，直到队列有新消息进来才立刻唤醒返回，性能极高，远胜for死循环去等
*/
// rdb.BLPop(ctx, timeout, keys...)
res, err := rdb.BLPop(ctx, 5*time.Second, "task:queue").Result()
if err == nil {
  fmt.Println("消费到了任务", res[1])
}


// Set 集合-去重且无序 
// 向集合添加元素 重复添加返回 0
// rdb.SAdd(ctx, key, members...)
count, err := rdb.SAdd(ctx, "post:like:888", "userA", "userB", "userC", "userA")
// count 此时为 2

// 判断是否在集合中 返回 true/false
// rdb.SIsMember(ctx, key, member)
isIn, err := rdb.SIsMember(ctx, "post:like:888", "userA").Result()
if isIn {
  // 已经在集合中了
}

// 获取集合中所有元素
// rdb.SMembers(ctx, key)
users, err := rdb.SMembers(ctx, "post:like:888").Result()


// ZSet （有序集合-每一个元素带一个分数， 常用于排行榜）
// 添加元素与分数
// rdb.ZSet(ctx, key, &redis.Z{Score: float64, Member: any})
err := rdb.ZSet(ctx, "game:rank", &redis.Z{
  Score: 999.9,
  Member: "player_alex",
})

// 获取前N的排行（降序）
// rdb.ZRevRangeWithScores(ctx, key, start, stop)
// 0-9 前10
rankList, err := rdb.ZRevRangeWithScores(ctx, "game:rank", 0, 9).Result()
if err != nil {
  return err
}


// 批量管道 Pipeline
/*
  若需要在同一个接口需要时间执行多次redis操作，会产生多次网络I/O往返延迟(RTT) 此时用pipeline
*/
func BatchCacheFields(ctx context.Context, rdb *redis.Client) error {
  // 1. 开启一个轻量的管道句柄
  pipe := rdb.Pipeline()

  // 2. 将要执行的命令放入管道
  cmd1 := pipe.Set(ctx, "key:1", "val:1", 10*time.Minute)
  cmd2 := pipe.Set(ctx, "key:2", "val:2", 10*time.Minute)
  cmd3 := pipe.Set(ctx, "key:3", "val:3", 10*time.Minute)
  cmd4 := pipe.Get(ctx, "key:4", "val:4")

  // 3. 一次提交
  _, err := pipe.Exec(ctx)
  if err := nil {
    return err
  }

  // 4. 读取合并回来的结果
  val4, _ := cmd3.Result()
  fmt.Println("批量管道中读取的key4 值为：", val4)
}

// 执行 lua 脚本 保证原子操作 ， 如： 扣减
// 1. 直接使用 Eval
script := `
  local = balance = tonumber(redis.call("GET", KEYS[1]))
  local amount = tonumber(ARGV[1])
  if balance < amount then 
    return 0
  end
  redis.call("DECRBY", KEYS[1], amount)
  return 1
`
result, err := rdb.Eval(
  ctx,
  script,
  []string{"wallet:1"},
  100
).Int()

// 2. 推荐：NewScript, 执行后会缓存脚本 sha1 -> evalsha, 不会重复发送整个 lua, 性能更好
/*
  script.Run(
    ctx, // ctx
    rdb, // rdb
    []string{"wallet:1"}, // keyArray []string{"wallet:1", "level:1"}
    100 // values ...
  ).Int()
*/
var deductScript = redis.NewScript(`
local = balance = tonumber(redis.call("GET", KEYS[1]))
local amount = tonumber(ARGV[1])

if not balance then
  return -1
end  

if balance < amount then 
  return 0
end
redis.call("DECRBY", KEYS[1], amount)
return 1
`)
ok, err := deductScript.Run(ctx, rdb, []string{"wallet:1"}, 100).Int()
switch ok {
  case -1:
  // 钱包不存在
  case 0:
  // 余额不足
  case 1:
  // 扣成功 
}

// 返回json
var script = redis.NewScript(`
  local cjson = cjson

  return cjson.encode({
    code=0,
    balance=100
  })
`)
type Resp struct {
  Code int `json:"code"`
  Balance int `json:"balance"`
}
var r Resp
str, _ := script.Run(ctx, rdb, ...).Text()
json.Unmarshal([]byte(str), &r)
```