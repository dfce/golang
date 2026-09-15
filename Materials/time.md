┌───────────────────────────────┐
│          时间获取              │
├───────────────────────────────┤
│ time.Now()         → 本地时间  │
│ time.Now().UTC()   → UTC时间  │
│ t.Unix()           → 秒        │
│ t.UnixNano()       → 纳秒      │
│ t.UnixNano()/1e6   → 毫秒      │
└───────────────────────────────┘

┌───────────────────────────────┐
│          时间组件              │
├───────────────────────────────┤
│ t.Year(), t.Month(), t.Day()  │
│ t.Hour(), t.Minute(), t.Second() │
│ t.Nanosecond()                │
│ t.Weekday()                   │
└───────────────────────────────┘

┌───────────────────────────────┐
│          时间比较              │
├───────────────────────────────┤
│ t1.Before(t2)                 │
│ t1.After(t2)                  │
│ t1.Equal(t2)                  │
└───────────────────────────────┘

┌───────────────────────────────┐
│          时间差计算            │
├───────────────────────────────┤
│ d := t2.Sub(t1) → time.Duration│
│ d.Hours(), d.Minutes(),       │
│ d.Seconds(), d.Milliseconds() │
└───────────────────────────────┘

┌───────────────────────────────┐
│          时间加减              │
├───────────────────────────────┤
│ t.Add(time.Hour*2 + ...)      │
│ t.AddDate(years, months, days)│
└───────────────────────────────┘

┌───────────────────────────────┐
│          格式化/解析           │
├───────────────────────────────┤
│ t.Format("2006-01-02 15:04:05")│
│ t.Format("2006-01-02")         │
│ time.Parse(layout, value)      │
│ time.ParseInLocation(layout, value, loc) │
└───────────────────────────────┘

┌───────────────────────────────┐
│          定时器/睡眠           │
├───────────────────────────────┤
│ time.Sleep(duration)           │
│ timer := time.NewTimer(dur)    │
│ <-timer.C                      │
│ ticker := time.NewTicker(dur)  │
│ <-ticker.C (循环触发)          │
└───────────────────────────────┘

┌───────────────────────────────┐
│          时区转换              │
├───────────────────────────────┤
│ loc, _ := time.LoadLocation("Asia/Tokyo") │
│ t.In(loc)       → 转换到 loc  │
│ t.Local()       → 系统本地时区│
│ t.UTC()         → UTC         │
└───────────────────────────────┘

┌───────────────────────────────┐
│          Duration              │
├───────────────────────────────┤
│ d := 90 * time.Minute          │
│ d.Hours(), d.Minutes(), d.Seconds(), d.Milliseconds() │
└───────────────────────────────┘

┌───────────────────────────────┐
│          其他                  │
├───────────────────────────────┤
│ t.IsZero() → 是否零值          │
│ sort.Slice(times, func(i,j int){return t[i].Before(t[j])})│
└───────────────────────────────┘


# 时间相关速查
                    ┌───────────────┐
                    │   时间获取     │
                    └───────────────┘
                          │
          ┌───────────────┼───────────────┐
          │                               │
     time.Now()                       time.Now().UTC()
          │                               │
      t.Unix() → 秒                  t.UnixNano() → 纳秒
          │
      t.UnixNano()/1e6 → 毫秒

                          │
                    ┌───────────────┐
                    │   时间组件     │
                    └───────────────┘
                          │
      ┌───────┬───────┬───────┬───────┬───────┐
      Year()  Month()  Day()  Hour()  Minute() Second()
      Nanosecond() Weekday()

                          │
                    ┌───────────────┐
                    │   时间比较     │
                    └───────────────┘
          ┌─────────┬─────────┬─────────┐
       t1.Before(t2) t1.After(t2) t1.Equal(t2)

                          │
                    ┌───────────────┐
                    │   时间差       │
                    └───────────────┘
                d := t2.Sub(t1) → time.Duration
          ┌─────────┬─────────┬─────────┐
          d.Hours() d.Minutes() d.Seconds()
          d.Milliseconds() d.Nanoseconds()

                          │
                    ┌───────────────┐
                    │   时间加减     │
                    └───────────────┘
          ┌─────────┬─────────┐
       t.Add(duration)  t.AddDate(years, months, days)

                          │
                    ┌───────────────┐
                    │ 格式化 / 解析 │
                    └───────────────┘
          t.Format("2006-01-02 15:04:05")
          time.Parse(layout, value)
          time.ParseInLocation(layout, value, loc)

                          │
                    ┌───────────────┐
                    │ 定时器 / 睡眠 │
                    └───────────────┘
          time.Sleep(duration)
          timer := time.NewTimer(d)
          <-timer.C
          ticker := time.NewTicker(d)
          <-ticker.C → 循环触发

                          │
                    ┌───────────────┐
                    │    时区       │
                    └───────────────┘
          loc, _ := time.LoadLocation("Asia/Tokyo")
          t.In(loc)   → 转换到 loc
          t.Local()   → 系统本地
          t.UTC()     → UTC

                          │
                    ┌───────────────┐
                    │   Duration    │
                    └───────────────┘
          d := 90 * time.Minute
          d.Hours() d.Minutes() d.Seconds() d.Milliseconds()

                          │
                    ┌───────────────┐
                    │     其他      │
                    └───────────────┘
          t.IsZero() → 是否零值
          sort.Slice(times, func(i,j int){ return t[i].Before(t[j]) })
