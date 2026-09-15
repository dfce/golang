package cases

import "fmt"

func MapCase() {
	nilTest()
}

func nilTest() {
	m := map[string]int{}
	m["a"] = 1
	fmt.Println(m)

	var m1 map[string]int
	if len(m1) == 0 {
		// if m1 == nil {
		m1 = make(map[string]int)
	}
	m1["B"] = 2
	fmt.Println(m1)
}
