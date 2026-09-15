package cases

import "fmt"

func BasicCase() {
	// varDeclarCase()
	// constCase()
	// structCase()
	// newCase()
	makeCase()
}

func varDeclarCase() {
	// 通过var声明并赋值
	// var b, c int float32 // 错误 ❌ Go 不允许在一行中为多个变量同时指定多个类型（除非它们类型相同）
	var i int = 100
	var i1, i2 int
	var f float32
	i1 = 2
	i2 = 3
	f = 3.14
	fmt.Println("变量声明：", i, i1, i2, f)
	fmt.Println("-----------------------")

	// 数组
	var arr = [3]int{1, 2, 3}
	arr1 := [...]int{3, 4, 5}
	var arr2 [3]int
	arr2[1], arr2[2] = 6, 7
	fmt.Println("变量数组：", arr, arr1, arr2)
	fmt.Println("-----------------------")

	// 指针
	var intPtr *int
	intPtr = &i
	fmt.Println(intPtr, &i) // intPtr == &i
	fmt.Printf("intPtr: %p, %p\n", intPtr, &i)
	fmt.Println("-----------------------")

	// 接口类型(空接口可以接收任何类型)
	var inter interface{}
	inter = i
	fmt.Println("interface 接收int:：", inter)
	inter = arr1
	fmt.Println("interface 接收数组:：", inter)
	fmt.Println("-----------------------")
}

func constCase() {
	const (
		B = 1 << (10 * iota)
		KB
		MB
		_ // GB
		TB
	)
	fmt.Println("B:", B)
	fmt.Println("KB:", KB)
	fmt.Println("MB:", MB)
	fmt.Println("TB:", TB)
}

type addres struct {
	province string
	city     string
}
type User struct {
	Name string
	Age  uint
	Addr addres
}

func structCase() {
	// 值类型
	u := User{
		Name: "test",
		Age:  18,
		Addr: addres{city: "cd"},
	}
	u.Addr.city = "wj"
	fmt.Println(u)

	// 指针类型
	u1 := &User{
		Name: "puser",
		Age:  20,
		Addr: addres{city: "cd.wj"},
	}
	u1.Addr.city = "cd"
	fmt.Println(u1)
}

/*
通过new 函数可以创建任意类型，并返回指针
*/
func newCase() {
	num := new(int)

	*num = 1
	fmt.Println(num, *num)

	// map
	mpPtr := new(map[string]*User)
	if *mpPtr == nil {
		fmt.Println("map 值为空")
	}
	fmt.Println(fmt.Sprintf("pointer: %p; %v", mpPtr, mpPtr))

	*mpPtr = make(map[string]*User)
	(*mpPtr)["alice"] = &User{
		Name: "初始化后再赋值",
	}
	fmt.Println(*mpPtr, mpPtr)
	fmt.Println((*mpPtr)["alice"])

	// slice
	slicePtr := new([]User)
	if *slicePtr == nil {
		fmt.Println("切片 值为空")
	}

	*slicePtr = append(*slicePtr, User{Name: "slcie,切片"})
	fmt.Println(*slicePtr)
	fmt.Println((*slicePtr)[0].Name)
}

// make 仅用于 切片、集合、通道的初始化
func makeCase() {
	// 初始化切片, 并设置长度和容量
	slice := make([]int, 10, 20)
	slice[0] = 10
	fmt.Println("slice:", slice)

	// 初始化集合, 并设置集合的初始大小
	mp := make(map[string]string, 10)
	mp["A"] = "a"
	fmt.Println("mp:", mp, mp["A"])

	// 初始化通道, 设置通道的读写方向和缓冲大小
	ch := make(chan int, 10)
	chWrite := make(chan<- int, 10)
	chRead := make(<-chan string, 10)

	fmt.Println("ch:", ch, chWrite, chRead)
}
