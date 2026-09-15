package stream

import (
	"bytes"
	"mini-srs/internal/protocol/flv"
	"mini-srs/internal/protocol/rtmp"
	"sync"
)

// 实现GOP缓存与序列头缓存(SRS的核心特性，能让Consumer进来秒开画面，不需要等待下一个I帧)

type GoCache struct {
	mu          sync.RWMutex
	vSeqHeader  *rtmp.Message
	aSeqHeader  *rtmp.Message
	gopDatabase []*rtmp.Message // 动态视频/音频混合GOP库
	hasIFrame   bool            // 标记当前 GOP组中是否已成功捕获到了关键I帧
}

func NewGoCache() *GoCache {
	return &GoCache{gopDatabase: make([]*rtmp.Message, 0)}
}

func (c *GoCache) Cache(msg *rtmp.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 彻底进行底层字节数组的深拷贝(Deep Copy),断绝一切内存复用污染
	srcBytes := msg.Payload.Bytes()
	if len(srcBytes) == 0 {
		return
	}
	// 申请全新的、独立的物理内存空间
	dstBytes := make([]byte, len(srcBytes))
	copy(dstBytes, srcBytes)

	cloned := &rtmp.Message{
		TypeID:    msg.TypeID,
		Length:    msg.Length,
		Timestamp: msg.Timestamp,
		StreamID:  msg.StreamID,
		Payload:   bytes.NewBuffer(dstBytes),
	}

	switch msg.TypeID {
	case 9: // 视频
		if flv.IsAVCSequenceHeader(srcBytes) {
			c.vSeqHeader = cloned
			return
		}
		if flv.IsKeyFrame(srcBytes) {
			// 规范1：遇到新视频I帧，彻底情况旧GOP库，重新开始累积当前视频组
			c.gopDatabase = c.gopDatabase[:0]
			c.hasIFrame = true
			c.gopDatabase = append(c.gopDatabase, cloned)
		} else {
			// 如果是普通的 P、B帧， 必须在当前GOP已经有 I 帧（解码参考基准）的前提下才收录
			if c.hasIFrame {
				c.gopDatabase = append(c.gopDatabase, cloned)
			}
		}

	case 8: // 音频
		if flv.IsAACSequenceHeader(srcBytes) {
			c.aSeqHeader = cloned
			return
		}
		// 规范2：如果当前GOP连第一个I帧都还没有，不要盲目的收录孤立的音频帧， 否则会让播放器时钟彻底换乱卡死
		if c.hasIFrame {
			c.gopDatabase = append(c.gopDatabase, cloned)
		}
	}
}

func (c *GoCache) SendToConsumer(sendFunc func(*rtmp.Message)) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 1. 首先下发最基础的解码配置序列头(SPS/PPS以及 AAC AudioSpecifiConfig)
	if c.vSeqHeader != nil {
		sendFunc(c.vSeqHeader)
		// // 发送时同样做一层防御性克隆，防止多用户并发读取同一Buffer 改变读写指针
		// sendFunc(&rtmp.Message{
		// 	TypeID:    c.vSeqHeader.TypeID,
		// 	Length:    c.vSeqHeader.Length,
		// 	Timestamp: 0, //初始头时间戳归0 //c.vSeqHeader.Timestamp,
		// 	StreamID:  c.vSeqHeader.StreamID,
		// 	Payload:   bytes.NewBuffer(c.vSeqHeader.Payload.Bytes()),
		// })
	}

	if c.aSeqHeader != nil {
		sendFunc(c.aSeqHeader)
		// sendFunc(&rtmp.Message{
		// 	TypeID:    c.aSeqHeader.TypeID,
		// 	Length:    c.aSeqHeader.Length,
		// 	Timestamp: 0, //初始头时间戳归0
		// 	StreamID:  c.aSeqHeader.StreamID,
		// 	Payload:   bytes.NewBuffer(c.aSeqHeader.Payload.Bytes()),
		// })
	}

	if len(c.gopDatabase) == 0 {
		return
	}

	// 2. 保持 FFmpeg 原生推送的完美音视频交织顺序直接全量下发（不盲目进行全量时间戳重排）
	// 因为前面已经实施了 I帧强对齐拦截，此时第一帧必定是I帧
	// 测地消灭 DTS out of oder 逆序错误
	for _, msg := range c.gopDatabase {
		sendFunc(&rtmp.Message{
			TypeID:    msg.TypeID,
			Length:    msg.Length,
			Timestamp: msg.Timestamp,
			StreamID:  msg.StreamID,
			Payload:   bytes.NewBuffer(msg.Payload.Bytes()),
		})
	}
}
