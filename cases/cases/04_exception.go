package cases

import "fmt"

/*
#
# 应用场景
1. 资源释放
2. 异常捕获

## defer
1. defer 关键词用来声明一个延迟调用函数, 该函数可以是匿名函数也可以是具名函数
2. defer 延迟函数执行时间(位置), 方法return之后, 返回参数到调用方法之前
3. defer 延迟函数可以在方法返回之后改变函数的返回值
4. 在方法结束(正常返回,异常结束)都会去调用defer声明的延迟函数, 可以有效避免因异常导致的资源无法释放的问题
5. 可以指定多个defer 延迟函数, 多个延时函数执行顺序为后进先出
6. defer 通常用于资源释放、异常捕获等场景, 例如: 关闭链接、文件等
7. defer 与 recover 配合可以实现异常捕获与处理逻辑
8. 不建议在for循环中使用

## recover
1. Go语言的内建函数, 可以让进入宕机流程中的goroutine恢复过来
2. recover 仅在延迟函数 defer中有效, 在正常执行过程中, 调用 recover 会返回nil 并且没有其他任何效果
3. 如果当前goroutine出现panic, 调用recover 可以捕获到 panic的输入值, 并且恢复正常执行

## panic
1. Go语言的一种异常机制
2. 可通过panic 函数主动抛出异常
*/

// defer 参数预计算
func DeferCase() {
	i := 1
	fmt.Println("start i:", i)

	defer func() {
		fmt.Println("defer 匿名函数1: 打印最终时的i:", i)
	}()

	defer func(j int) {
		fmt.Println("defer 匿名函数2: 传入1输出2 j:", j)
	}(i + 1)

	i = 99
	fmt.Println("end i:", i)
}
