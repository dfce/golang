package cases

import "fmt"

func StructCase() {
	// testNilStruct()
	test()
}

// 测试空struct
type Info struct {
	Id    int
	Name  string
	Other map[string]string
}

func testNilStruct() {
	info := &Info{}
	fmt.Println(info, *info)

	var i1 *Info

	fmt.Println(i1, i1 == nil, &i1)
}

type mstr string

func test() {
	var a []mstr
	setA(&a)
	fmt.Println(a)
}
func setA(s *[]mstr) {
	*s = []mstr{"a", "b"}
}
