package rtmp

import (
	"encoding/binary"
	"io"
)

/*
	Chunk 编码层。下发数据给播放器、或做信令响应时，将大报文切片发送
*/

type ChunkWriter struct {
	w            io.Writer
	outChunkSize uint32
}

func NewChunkWriter(w io.Writer) *ChunkWriter {
	return &ChunkWriter{w: w, outChunkSize: 128}
}

func (cw *ChunkWriter) SetChunkSize(size uint32) {
	cw.outChunkSize = size
}

func (cw *ChunkWriter) WriteMessage(msg *Message) error {
	if msg.Payload == nil {
		return nil
	}
	data := msg.Payload.Bytes()
	offset := uint32(0)
	total := uint32(len(data))
	if total == 0 {
		return nil
	}

	// 【强标准对齐】：根据RTMP规范，严格划分底层CSID(Chunk Stream ID)
	csid := uint8(3) // 默认控制/信令/媒体常用的 CSID
	switch msg.TypeID {
	case 4:
		csid = 2
	case 8, 9:
		csid = 4
	}

	for offset < total {
		if offset == 0 {
			// 1. 写入Basic Header(fmt = 0, 表示完整大头)
			if _, err := cw.w.Write([]byte{0x00 | csid}); err != nil {
				return err
			}
			// 2. 构造标准的11字节Message Header
			h := make([]byte, 11)

			// 3. 时间戳(3 字节， 大端序)
			ts := msg.Timestamp
			if ts >= 0xFFFFFF {
				ts = 0xFFFFFF // 超过则触发扩展时间戳（此处做标准截断防御）
			}
			h[0] = byte(ts >> 16)
			h[1] = byte(ts >> 8)
			h[2] = byte(ts)

			// 4. 载荷长度(3 字节， 大端序)
			h[3] = byte(total >> 16)
			h[4] = byte(total >> 8)
			h[5] = byte(total)

			// 5. 消息类型ID(1 字节)
			h[6] = msg.TypeID

			// 6. 流ID(4 字节)
			binary.LittleEndian.PutUint32(h[7:11], msg.StreamID)

			if _, err := cw.w.Write(h); err != nil {
				return err
			}
			// 如果触发了扩展时间戳，按规范在此处追加4字节大端时间戳
			if ts == 0xFFFFFF {
				extTs := make([]byte, 4)
				binary.BigEndian.PutUint32(extTs, msg.Timestamp)
				if _, err := cw.w.Write(extTs); err != nil {
					return err
				}
			}
		} else {
			// 7. 后续分片：写入Type 3简略包头(fmt=3,复用之前的所有Header属性)
			if _, err := cw.w.Write([]byte{0xC0 | csid}); err != nil {
				return err
			}
			// 如果触发了扩展时间戳，按规范在此处追加4字节大端时间戳
			if msg.Timestamp >= 0xFFFFFF {
				extTs := make([]byte, 4)
				binary.BigEndian.PutUint32(extTs, msg.Timestamp)
				if _, err := cw.w.Write(extTs); err != nil {
					return err
				}
			}
		}

		// 8. 计算本次切片发送的安全边界
		size := min(total-offset, cw.outChunkSize)

		if _, err := cw.w.Write(data[offset : offset+size]); err != nil {
			return err
		}

		offset += size
	}

	return nil
}
