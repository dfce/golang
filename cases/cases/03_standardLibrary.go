package cases

import (
	"cmp"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"sort"
)

/*
标准库
*/

func StandardLibCase() {
	// encoding()
	// regex()
	// sortCase()
	testSort()
}

func encoding() {
	type user struct {
		Id   int64
		Name string
		Age  uint
	}

	u := user{Id: 1, Name: "text", Age: 19}
	// encoding
	bytes, err := json.Marshal(u)
	if err == nil {
		fmt.Println("encoding:", bytes)
	} else {
		fmt.Println("encoding err:", err)
	}

	// decoding
	u1 := user{}
	err = json.Unmarshal(bytes, &u1)
	fmt.Println(u1, err)
}

func regex() {
	reg := regexp.MustCompile(`^[a-z]+\[[0-9]+\]$`)

	fmt.Println(reg.MatchString("abuc[1234]"))
	fmt.Println(reg.MatchString("abuc[1234a]"))
	fmt.Println(reg.MatchString("1abuc[1234a]"))

	bytes := reg.FindAll([]byte("abuc[1234]"), -1)
	fmt.Println(string(bytes[0]))

	//
	// fmt.Println(reg.Match([]byte("abuc[1234as]")))
	bytes1 := reg.FindAll([]byte("abuc[1234as]"), -1)
	fmt.Println("len(bytes1):", len(bytes1))
	if len(bytes1) > 0 {
		fmt.Println(string(bytes1[0]))
	}
}

// sort
type sortData struct {
	Id   uint64
	Name string
	Age  uint8
}
type ById []sortData

func (a ById) Len() int {
	return len(a)
}
func (a ById) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ById) Less(i, j int) bool {
	return a[i].Id < a[j].Id
}
func sortCase() {
	list := []sortData{
		{Id: 1, Name: "Name-1", Age: 8},
		{Id: 10, Name: "Name-10", Age: 18},
		{Id: 9, Name: "Name-9", Age: 12},
		{Id: 7, Name: "Name-7", Age: 10},
		{Id: 3, Name: "Name-3", Age: 6},
		{Id: 5, Name: "Name-5", Age: 82},
		{Id: 21, Name: "Name-21", Age: 21},
	}

	sort.Sort(ById(list))
	list2 := ById{
		{Id: 1, Name: "Name-1", Age: 8},
		{Id: 10, Name: "Name-10", Age: 18},
		{Id: 9, Name: "Name-9", Age: 12},
		{Id: 7, Name: "Name-7", Age: 10},
		{Id: 3, Name: "Name-3", Age: 6},
		{Id: 5, Name: "Name-5", Age: 82},
		{Id: 21, Name: "Name-21", Age: 21},
	}
	sort.Sort(list2)
	fmt.Println("list2:", list2)

	// Go 1.21+ 泛型排序（slices 包）
	// 自动默认排序
	strs := []string{"banana", "apple", "cherry"}
	slices.Sort(strs)
	fmt.Println(strs)

	// 自定义结构体排序
	// 顺序
	slices.SortFunc(list, func(a, b sortData) int {
		return cmp.Compare(a.Age, b.Age)
	})
	fmt.Println(list)
	// 倒序
	slices.SortFunc(list, func(a, b sortData) int {
		return cmp.Compare(b.Id, a.Id)
	})
	fmt.Println(list)
}

type Persion struct {
	Name string
	Age  int
}
type PersionSlice []Persion

// 实现三个标准接口方法
func (p PersionSlice) Len() int           { return len(p) }
func (p PersionSlice) Less(i, j int) bool { return p[i].Age < p[j].Age }
func (p PersionSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// sort 都是就地排序(In-place) Go的排序函数都会直接修改原切片的内存数据
// 不会返回一个新切片。 如果需要保留原数组，需要在排序前 copy 后sort
func testSort() {
	p2 := PersionSlice{{"D", 50}, {"C", 20}, {"E", 30}}
	p3 := make(PersionSlice, len(p2), cap(p2))
	copy(p3, p2)
	slices.SortFunc(p2, func(a, b Persion) int {
		return cmp.Compare(a.Age, b.Age)
	})
	fmt.Println("p2", p2)
	fmt.Println("p3", p3)

	fmt.Println("----------- sort in-place -----------")
	people := PersionSlice{{"C", 50}, {"A", 20}, {"B", 30}}
	var p1 PersionSlice = make(PersionSlice, len(people), cap(people))
	copy(p1, people)
	sort.Sort(people)
	fmt.Println("people", people)
	fmt.Println("p1", p1)
}
