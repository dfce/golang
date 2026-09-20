package middleware

import (
	"bytes"
	"encoding/json"
	"go-module/internal/config"
	"go-module/pkg/constant"
	"go-module/pkg/util"
	"io"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UploadFile struct {
	Fieldname string `json:"fieldName"`
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	MIMEType  string `json:"mimeType,omitempty"`
}
type MultipartInfo struct {
	Values url.Values
	Files  []UploadFile
}
type ReqInfo struct {
	Query     string          `json:"query,omitempty"`
	Params    gin.Params      `json:"params,omitempty"`
	Body      json.RawMessage `json:"body,omitempty"`      // application/json // []byte `json:"body,omitempty"`
	Form      url.Values      `json:"form,omitempty"`      // x-www-form-urlencoded
	Multipart *MultipartInfo  `json:"multipart,omitempty"` // multipart/form-data
}

func RequestInfo(logger *zap.Logger, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.App.ENV == "dev" {
			printfRequestInfo(c, logger)
		}

		c.Next()
	}
}

func printfRequestInfo(c *gin.Context, logger *zap.Logger) {
	info := &ReqInfo{} // var info *ReqInfo
	hasReqInfo := false

	if c.Request.URL.RawQuery != "" {
		hasReqInfo = true
		info.Query = c.Request.URL.RawQuery
	}
	if c.Params != nil {
		hasReqInfo = true
		info.Params = c.Params
	}
	getPostData(c, logger, info, &hasReqInfo)

	logInfo := []zap.Field{
		zap.String("method", c.Request.Method),         // 请求方法：GET, POST
		zap.String("path", c.Request.URL.Path),         // 路由路径
		zap.String(constant.TraceKey, util.TraceID(c)), // 链路追踪 ID
	}

	if hasReqInfo {
		logInfo = append(logInfo, zap.Any("request_info", info))
	}

	logger.Debug("", logInfo...)
}

func getPostData(c *gin.Context, logger *zap.Logger, info *ReqInfo, hasReqInfo *bool) {
	// c.Request.Header.Get("Content-Type")
	switch c.ContentType() {
	case "application/json":
		// 保存 Body
		// 读取body内容(body是 io.ReadCloser, 只能读一次)
		body, err := io.ReadAll(c.Request.Body)
		if err == nil && len(body) > 0 {
			// 读完后恢复 c.Request.Body, 让后续仍可读
			// 恢复 Body，后续 BindJSON 还能继续读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			info.Body = json.RawMessage(body)
			// 如果 info.Body 设置为 []byte , info.Body = body
			*hasReqInfo = true
		}
	case "application/x-www-form-urlencoded":
		// 保存 Form (复制，否则后续修改，这里也会同步)
		if err := c.Request.ParseForm(); err != nil {
			logger.Warn("parse form failed", zap.Error(err))
			return
		} else {
			info.Form = cloneFormValues(c.Request.PostForm)
			*hasReqInfo = true
		}
	case "multipart/form-data":
		// 保存 Form
		err := c.Request.ParseMultipartForm(32 << 20) // 限制最大 32MB
		if err != nil {
			logger.Warn("parse multipart failed", zap.Error(err))
			return
		}
		form := c.Request.MultipartForm
		if form == nil {
			break
		}

		info.Multipart = &MultipartInfo{}
		info.Multipart.Values = cloneFormValues(form.Value)
		fcount := 0
		fstop := false
		for fieldName, headers := range form.File {
			for _, h := range headers {
				fcount++
				if fcount > 20 { // 最多上传文件数
					fstop = true
					break
				}
				info.Multipart.Files = append(info.Multipart.Files, UploadFile{
					Fieldname: fieldName,
					Filename:  h.Filename,
					Size:      h.Size,
					MIMEType:  h.Header.Get("Content-Type"),
				})
			}
			if fstop {
				break
			}
		}
		*hasReqInfo = true
	}
}

func cloneFormValues(values url.Values) url.Values {
	cloned := make(url.Values, len(values))

	for k, v := range values {
		if isSensitiveField(k) {
			cloned[k] = []string{"[REDACTED]"}
			continue
		}
		cloned[k] = append([]string(nil), v...)
	}
	return cloned
}

func isSensitiveField(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "password", "passwd", "pwd", "confirm", "confirmation", "token", "secret":
		return true
	default:
		return false
	}
}
