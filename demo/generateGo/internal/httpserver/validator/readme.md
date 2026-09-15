# 参数绑定 校验

```go
package myvalidator

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)


// 范型封装
// 获取指定类型的参数， 并绑定
// 直接用 gin 自定义的 binding.Binding 对象作为参数
/*
	调用方式：
	req, err := GetParamGeneric[UserLogin](c, binding.JSON)
*/
func GetParamGeneric[T any](c *gin.Context, btyte binding.Binding) (*T, error) {
	var data T // 自动创建对应的结构体实例
	var err error

	if err := c.ShouldBindWith(&data, btyte); err != nil {
		return nil, err
	}
	return &data, err
}

/*
调用方式：
req, err := GetParamGeneric[UserLogin](c, "json")
*/
func GetParamGenericBtype[T any](c *gin.Context, btyte string) (*T, error) {
	var data T // 自动创建对应的结构体实例
	var err error

	switch btyte {
	case "json":
		err = c.ShouldBindJSON(&data)
	case "queyr":
		err = c.ShouldBindQuery(&data)
	case "uri":
		err = c.ShouldBindUri(&data)
	case "form":
		err = c.ShouldBind(&data)
	default:
		err = errors.New("invalid binging type")
	}
	if err != nil {
		return nil, err
	}
	return &data, err
}

// 使用 gin 自定义的 binding.Binding 对象作为参数
/*
	调用方式：
	var req UserLogin
	req, err := GetparamElegant(c, &req, binding.JSON)
*/
func GetparamElegant(c *gin.Context, obj any, b binding.Binding) error {
	return c.ShouldBindWith(obj, b)
}

```