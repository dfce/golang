package demo

// HealthStatusRes Demo 就绪检查响应数据。
type HealthStatusRes struct {
	// 应用状态。
	Status string `json:"status" example:"ok"`
	// Redis 依赖状态。
	Redis CheckResult `json:"redis"`
}

// CheckResult 外部依赖检查结果。
type CheckResult struct {
	// 是否启用该依赖。
	Enabled bool `json:"enabled" example:"true"`
	// 依赖是否可用。
	OK bool `json:"ok" example:"true"`
	// 检查失败时的错误信息。
	Error string `json:"error,omitempty" example:"connection refused"`
}
