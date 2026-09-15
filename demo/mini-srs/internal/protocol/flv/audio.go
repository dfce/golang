package flv

func IsAACSequenceHeader(payload []byte) bool {
	if len(payload) < 2 {
		return false
	}
	// 第一个字节 4位为 SoundFormat (10=ACC),第二字节为 AACPacketType(0=AudioSpecificConfig)
	return (payload[0]>>4) == 10 && payload[1] == 0
}
