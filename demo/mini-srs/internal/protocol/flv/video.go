package flv

// 在RTMP流中提取并定义音视频元数据状态

func IsKeyFrame(payload []byte) bool {
	if len(payload) == 0 {
		return false
	}

	// RTMP AVC/H264 格式中，第一个字节高 4位为：Frame，1为 KeyFrame
	return (payload[0] >> 4) == 1
}

func IsAVCSequenceHeader(payload []byte) bool {
	if len(payload) < 2 {
		return false
	}
	// 第二个字节为 AVCPacketType，0为 AVC sequence header （SPS/PPS）
	return payload[1] == 0
}
