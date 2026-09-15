package rtmp

/*
	定义RTMP核心的消息结构体，它由多个Chunk重组而成
*/

import "bytes"

type Message struct {
	TypeID    uint8
	Length    uint32
	Timestamp uint32
	StreamID  uint32
	Payload   *bytes.Buffer
}

func NewMessage(typeID uint8, length, ts, streamID uint32) *Message {
	return &Message{
		TypeID:    typeID,
		Length:    length,
		Timestamp: ts,
		StreamID:  streamID,
		Payload:   bytes.NewBuffer(make([]byte, 0, length)),
	}
}
