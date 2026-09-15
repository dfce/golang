package cases

import "fmt"

func SliceTest() {
	var s1 []int
	fmt.Println(s1, s1 == nil) // true

	s2 := []int{}
	fmt.Println(s2, s2 == nil) // false

	s3 := make([]int, 0)
	fmt.Println(s3, s3 == nil) // false

	testUnshift()
}

func testUnshift() {
	// 提前尽量算预留容量， 避免多次重复扩容（重新分配扩容和数据拷贝。）
	// // 初始化一个空切片
	// var s2 []int
	// s2 := []int{}
	// var s2 = make([]int, 0)

	s1 := []int{2, 3, 4, 5, 6}
	s2 := []int{0, 1}
	s2 = append(s2, s1...)

	fmt.Println(s2)

	// 不重新分配内存的情况下向原来的slice 头部插入。cap 足够的情况下
	s := make([]int, 3, 10)
	s[0], s[1], s[2] = 3, 4, 5
	fmt.Println(s)
	s = append(s, 0)
	copy(s[1:], s[:len(s)-1])
	s[0] = 2
	fmt.Println(s)
}
