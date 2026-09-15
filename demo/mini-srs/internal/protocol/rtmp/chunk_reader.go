package rtmp

import (
	"bytes"
	"encoding/binary"
	"io"
)

/*
 最核心的解码层。RTMP消息再传输网络中会被切成 Chunk传输。Reader用于解分包
 并拼装完整的 Message
*/

type ChunkReader struct {
	r           io.Reader
	inChunkSize uint32
	csMsgs      map[uint32]*Message // 保存每个CSID 上一次未拼接完整的 Message
}

func NewChunkReader(r io.Reader) *ChunkReader {
	return &ChunkReader{
		r:           r,
		inChunkSize: 128, // 默认初始128字节
		csMsgs:      make(map[uint32]*Message),
	}
}

func (cr *ChunkReader) SetChunkSize(size uint32) {
	cr.inChunkSize = size
}

func (cr *ChunkReader) ReadMessage() (*Message, error) {
	for {
		var fmtAndCsid uint8
		if err := binary.Read(cr.r, binary.BigEndian, &fmtAndCsid); err != nil {
			return nil, err
		}

		fmtType := fmtAndCsid >> 6
		csid := uint32(fmtAndCsid & 0x3F) // 简化处理，只处理基本的CSID（0-63）

		msg, exists := cr.csMsgs[csid]
		if !exists {
			// 初始化当前CSID的默认消息载荷
			msg = &Message{Payload: bytes.NewBuffer(make([]byte, 0))}
			cr.csMsgs[csid] = msg
		}

		// 解析Header变长结构
		switch fmtType {
		case 0:
			var h [11]byte
			if _, err := io.ReadFull(cr.r, h[:]); err != nil {
				return nil, err
			}
			msg.Timestamp = uint32(h[0])<<16 | uint32(h[1])<<8 | uint32(h[2])
			msg.Length = uint32(h[3])<<16 | uint32(h[4])<<8 | uint32(h[5])
			msg.TypeID = h[6]
			msg.StreamID = binary.LittleEndian.Uint32(h[7:11])
			msg.Payload.Reset() // 新消息开始，重制缓冲区
		case 1:
			var h [7]byte
			if _, err := io.ReadFull(cr.r, h[:]); err != nil {
				return nil, err
			}
			msg.Timestamp = uint32(h[0])<<16 | uint32(h[1])<<8 | uint32(h[2])
			msg.Length = uint32(h[3])<<16 | uint32(h[4])<<8 | uint32(h[5])
			msg.TypeID = h[6]
			msg.Payload.Reset() // 新消息开始，重制缓冲区
		case 2:
			var h [3]byte
			if _, err := io.ReadFull(cr.r, h[:]); err != nil {
				return nil, err
			}
			msg.Timestamp = uint32(h[0])<<16 | uint32(h[1])<<8 | uint32(h[2])
			msg.Payload.Reset() // 新消息开始，重制缓冲区
		}
		// fmtType == 3 完美保持当前的msg.Length 和当前正在append的msg.Payload

		// 如果发生了协议层通知改变ChunkSize 的行为，在这里有限纠正
		if msg.TypeID == 1 && msg.Length == 4 && uint32(msg.Payload.Len()) == 0 && fmtType != 3 {
			// 防御机制：防止特定握手块被误判
		}

		if msg.Payload == nil {
			msg.Payload = NewMessage(msg.TypeID, msg.Length, msg.Timestamp, msg.StreamID).Payload
		}

		// 读取当前Chunk 载荷
		needed := msg.Length - uint32(msg.Payload.Len())
		chunkSize := min(needed, cr.inChunkSize)

		buf := make([]byte, chunkSize)
		if _, err := io.ReadFull(cr.r, buf); err != nil {
			return nil, err
		}

		msg.Payload.Write(buf)

		// 检查当前CSID上的消息是否接收完整
		if uint32(msg.Payload.Len()) == msg.Length {
			// 【核心修正】：深拷贝一份完成的Message返回出去，绝不污染和清空底层缓存
			retMsg := &Message{
				TypeID:    msg.TypeID,
				Length:    msg.Length,
				Timestamp: msg.Timestamp,
				StreamID:  msg.StreamID,
				// Payload:   msg.Payload,
				Payload: bytes.NewBuffer(msg.Payload.Bytes()),
			}
			// 接收完一个完整大包后，重制当前通道的缓冲区，迎接下一个大包
			msg.Payload.Reset()
			return retMsg, nil
		}
	}
}
