package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	validatorV10 "github.com/go-playground/validator/v10"
)

// 验证并绑定参数
func Validate[T any](c *gin.Context, b binding.Binding) (*T, string, error) {
	var data T
	var err error

	if err = c.ShouldBindWith(&data, b); err != nil {
		var errs validatorV10.ValidationErrors
		if errors.As(err, &errs) {
			var msgList []string
			t := reflect.TypeOf(data)

			for _, e := range errs {
				// 获取结构体字段 comment
				fieldName := e.Field()
				field, exists := t.FieldByName(fieldName)
				if exists {
					if comment := field.Tag.Get("comment"); comment != "" {
						fieldName = comment
					} else if jsonTag := field.Tag.Get("json"); jsonTag != "" {
						fieldName = strings.Split(jsonTag, ",")[0]
					}
				}

				// 追加：针对高频标签的响应
				var msg string
				switch e.Tag() {
				case "required":
					msg = fmt.Sprintf("[%s] 必填，不能为空", fieldName)
				case "min":
					msg = fmt.Sprintf("[%s] 长度不能小于 %s", fieldName, e.Param())
				case "max":
					msg = fmt.Sprintf("[%s] 长度不能大于 %s", fieldName, e.Param())
				case "gt":
					msg = fmt.Sprintf("[%s] 必须大于 %s", fieldName, e.Param())
				case "lt":
					msg = fmt.Sprintf("[%s] 必须小于 %s", fieldName, e.Param())
				case "len":
					msg = fmt.Sprintf("[%s] 长度必须等于 %s", fieldName, e.Param())
					// 格式校验
				case "email":
					msg = fmt.Sprintf("[%s] 格式不正确", fieldName)
				case "url":
					msg = fmt.Sprintf("[%s] 必须是一个合法的网址 URL", fieldName)
				case "uuid", "uuid4":
					msg = fmt.Sprintf("[%s] 必须是一个合法的 UUID 格式", fieldName)
				case "numeric":
					msg = fmt.Sprintf("[%s] 只能包含纯数字", fieldName)
				case "alphanum":
					msg = fmt.Sprintf("[%s] 只能包含字母和数字", fieldName)

				// 追加：日期与时间格式校验 (对应 binding:"datetime=2006-01-02")
				case "datetime":
					msg = fmt.Sprintf("[%s] 必须符合指定的日期时间格式: %s", fieldName, e.Param())

				// 追加：枚举校验 (对应 binding:"oneof=admin user guest")
				case "oneof":
					// 将空格替换为逗号，例如 "admin user" -> "admin, user"
					opts := strings.ReplaceAll(e.Param(), " ", ", ")
					msg = fmt.Sprintf("[%s] 必须是以下可选值之一: [%s]", fieldName, opts)

				// 追加：跨字段关联校验 (例如确认密码 对应 binding:"eqfield=Password")
				case "eqfield":
					// 尝试找到被关联字段的 comment 标签
					targetFieldName := e.Param()
					if targetField, ok := t.FieldByName(e.Param()); ok {
						if targetComment := targetField.Tag.Get("comment"); targetComment != "" {
							targetFieldName = targetComment
						}
					}
					msg = fmt.Sprintf("[%s] 必须与 [%s] 保持一致", fieldName, targetFieldName)

				// 兜底逻辑：处理其他未配置的规则
				default:
					msg = fmt.Sprintf("[%s] 格式校验未通过 (%s)", fieldName, e.Tag())
				}
				msgList = append(msgList, msg)
			}
			return nil, strings.Join(msgList, "; "), err
		}
		return nil, "数据格式解析失败, 请检查请求体", err
	}

	return &data, "", nil
}
