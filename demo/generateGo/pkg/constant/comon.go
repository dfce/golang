package constant

import "generatego/pkg/util"

/*
	存放全局共享用常量
*/

const (
	DateLayout = "2006-01-02"
	TimeLayout = "2006-01-02 15:04:05"
	TimeFormat = "2006-01-02 15:04:05.000"

	TraceName   = "trace_id"
	TraceHeader = "X-Trace-Id"

	JwtPrefix = "Bearer "
)

var (
	JwtKey    = []byte(util.GetEnv("JWT_SECRET_KEY", "config_secret_key"))
	JwtExpire = util.IntEnv("JWT_TOKEN_EXPIRY", 7200) // Second
)
