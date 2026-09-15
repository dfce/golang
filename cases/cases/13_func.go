package cases

import "fmt"

func FuncCase() {
	optionFunc()
}

// 实现接口 可选参数
func optionFunc() {
	// 1
	// test1(1, 2, 3, 4, 5)

	// // 2
	// test2(1, 2, Opt{skip: true})
	// test2(4, 5, Opt{})
	// test2(4, 5, Opt{})

	// 3
	test3(3, 4)
	test3(3, 4, WithSkip())
}

func test1(a, b int, c ...int) {
	fmt.Println("可选参数 ", c)
	fmt.Println(a + b)
}

// Option Pattern
type Opt struct {
	skip  bool
	retry int
}

func test2(a, b int, opt Opt) {
	fmt.Println("Option Pattern", opt)
	fmt.Println(a + b)
}

// 3
type AuthConfig struct {
	Skip  bool
	Retry int
}
type AuthOption func(*AuthConfig)

func WithSkip() AuthOption {
	return func(c *AuthConfig) {
		c.Skip = true
	}
}
func test3(a, b int, opts ...AuthOption) {
	cfg := AuthConfig{} // 默认

	// 如果有可选参数 ， 设置cfg
	for _, opt := range opts {
		opt(&cfg)
	}

	fmt.Println(a+b, cfg.Skip)
}
