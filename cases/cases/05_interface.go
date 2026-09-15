package cases

import "fmt"

/*
interface “鸭子类型”(Duck Typing)
		“如果一个动物走起来像鸭子，叫起来也像鸭子，那它就是鸭子”
只要一个结构体实现了接口中定义的所有方法，编译器就会自动认为该结构体实现了该接口，不需要任何显式声明。
*/

func MainCase() {
	// 1.基础示例：定义与隐式实现
	file := LocalFile{Path: "../readme.md"}
	PrintContent(file)

	oss := OssFile{Bucket: "test-bucket"}
	PrintContent(oss)

	// 2. 空接口 interface{} 与 any

}

// 基础示例：定义与隐式实现
// 1. 定义接口：包含一个Read方法
type Reader interface {
	Read() string
}

// 2. 结构体A：本地文件
type LocalFile struct {
	Path string
}

// LocalFile 绑定了Read 方法->自动实现了Reader接口
func (f LocalFile) Read() string {
	return "从本地路径" + f.Path + "读取数据"
}

// 3. 结构体B：阿里云OSS 远程文件
type OssFile struct {
	Bucket string
}

// OssFile 也绑定了Read 方法->自动实现了Reader接口
func (o OssFile) Read() string {
	return "从阿里云OSS存储桶" + o.Bucket + "读取数据"
}

// 4. 通用业务函数：只认接口，不认具体结构体
func PrintContent(r Reader) {
	// 多态：根据传入的实际对象，调用不同的Read实现
	fmt.Println(r.Read())
}

// 2. 空接口 interface{} 与 any
func ProcessData(i any) {
	// 使用 Type Switch 动态判断数据真实底层类型
	switch v := i.(type) {
	case int:
		fmt.Printf("整数，乘以2: %d\n", v*2)
	case string:
		fmt.Printf("字符串，长度为: %d\n", len(v))
	case LocalFile:
		fmt.Printf("文件对象，路径是: %s\n", v.Path)
	default:
		fmt.Println("未知类型")
	}
}
