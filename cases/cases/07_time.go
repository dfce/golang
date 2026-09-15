package cases

import (
	"fmt"
	"time"
)

/*
# time 包整体结构
 ```
 time
 ├── Time          // 时间对象
 ├── Duration      // 时间间隔
 ├── Timer         // 一次性定时器
 ├── Ticker        // 周期定时器
 ├── After         // 延迟触发
 ├── Sleep         // 睡眠
 ├── Since         // 已过去多久
 ├── Until         // 距离多久
 ├── Parse         // 字符串转时间
 ├── Format        // 时间格式化
 ├── Unix          // 时间戳
 ├── Location      // 时区
 ```
*/

func CaseTime() {
	// getNow()
	// judgment()
	// duration()
	// calculate()
	// sinceUntil()
	// timer()
	// ticker()
	parse()
}

func getNow() {
	// 获取当前时间
	now := time.Now()
	fmt.Println(now) // 2026-07-02 16:23:15.123456 +0800 CST

	// 获取各种字段
	fmt.Println(now.Year())
	fmt.Println(now.Month())
	fmt.Println(now.Day())

	fmt.Println(now.Hour())
	fmt.Println(now.Minute())
	fmt.Println(now.Second())

	fmt.Println(now.Weekday())
	fmt.Println(now.YearDay())

}

// 判断
func judgment() {
	t1 := time.Now()
	t2 := t1.Add(time.Hour)

	fmt.Println(t1.Before(t2))
	fmt.Println(t1.After(t2))
	fmt.Println(t1.Equal(t2))
}

// 时间间隔 type Duration int64 单位：纳秒/ns
func duration() {
	// 常用单位：
	// time.Nanosecond  纳秒
	// time.Microsecond	微妙
	// time.Millisecond 毫秒
	// time.Second			秒
	// time.Minute 			分
	// time.Hour				时

	timeout := 5*time.Second + 300*time.Millisecond
	fmt.Println(timeout)
}

// 计算
func calculate() {
	// AddData 增加 年月日
	now := time.Now()
	future := now.Add(time.Hour)
	future = future.Add(-2 * time.Hour)
	du := future.Sub(now)
	_ = future.Add(du)
	_ = future.AddDate(1, 2, 3)

}

// 计算耗时
func sinceUntil() {
	start := time.Now()
	time.Sleep(time.Second)
	fmt.Println(time.Since(start))

	// 距离未来多久
	now := time.Now()
	expire := now.Add(time.Hour)
	fmt.Println("距离", now, time.Until(expire).Seconds())
	fmt.Println("距离", now, time.Until(expire))
	fmt.Println("距离", now, expire.Sub(now).Seconds())
	fmt.Println("距离", now, expire.Sub(now))
}

// timer
func timer() {
	// 一次
	timer := time.NewTimer(3 * time.Second)
	<-timer.C
	fmt.Println("3 timeout")

	// stop
	timer.Stop()
	// reset 重新开始
	timer.Reset(4 * time.Second)
	<-timer.C
	fmt.Println("4 timeout")

	/*
		After 其实就是 NewTimer()简写

		(等价于):
		timer1 := time.NewTimer(time.Second)
		<-timer1.C
	*/
	<-time.After(time.Second)
}

// Ticker 周期 ticker 必须 stop
func ticker() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	i := 0
	for range ticker.C {
		fmt.Println(time.Now())
		i++
		if i >= 10 {
			return
		}
	}
}

// AfterFunc
func afterFunc() {
	time.AfterFunc(time.Second, func() {
		fmt.Println("time.Second 后自动执行")
	})
}

// Parse
func parse() {
	fmt.Println(time.Now().UTC())
	fmt.Println(time.Now().Local())
}
