package cases

import (
	"cmp"
	"fmt"
)

/*
	泛型： 1.18引入generics
	1.	即开发过程中编写适用于所有类型的模版，只有在具体使用的时候才能确定其真正的类型

	作用、应用场景
	1.	增加代码复用率
	2.	不同类型间代码复用
	3.	增加了编译器的负担，降低编译效率

	泛型：(函数、类型、接口、结构体、interface)
	限制：
	1. 匿名结构体与函数不支持泛型
	2. 不支持类型断言
	3. 不支持泛型方法，只能通过receiver来实现方法的泛型处理
	4. ~后的类型必须为基本类型，不能为接口类型
*/

func GenericTCase() {
	getMaxTest()
}

// 比较数字大小， 不用泛型
// func MaxInt(a, b int) int {if a > b {return a}; return b}
// func MaxFloat(a, b float64) float64 {if a > b {return a}; return b}
// 不加 ~ 则会严格要求 int, ~int64 会包含 type myint int64, myint 也是合格的
type Numb interface {
	int | ~int64 | ~float32 | ~float64
}

func getMax[T Numb](a, b T) T {
	if a > b {
		return a
	}
	return b
}
func getMaxElegant[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func getMaxTest() {
	fmt.Println(getMax(1, 5))
	// 1.21+ 使用官方标准的 cmp.Ordered 约束

	fmt.Println(getMaxElegant("a", "m"))
	fmt.Println(getMaxElegant(3.14, 5.3))
}
