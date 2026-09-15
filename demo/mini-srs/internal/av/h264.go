package av

// 预留给未来做容器转换(如RTMP转HLS或者是协议过滤)的解析模块

type Nalu struct {
	Type uint8
	Data []byte
}

// 占位
func ParseH264(payload []byte) {}
