package stream

import (
	"bytes"
	"mini-srs/internal/protocol/flv"
	"mini-srs/internal/protocol/rtmp"
	"sort"
	"sync"
)

// 实现GOP缓存与序列头缓存(SRS的核心特性，能让Consumer进来秒开画面，不需要等待下一个I帧)

type GoCache struct {
	mu          sync.RWMutex
	vSeqHeader  *rtmp.Message
	aSeqHeader  *rtmp.Message
	gopDatabase []*rtmp.Message // 动态视频/音频混合GOP库
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
		}
		c.gopDatabase = append(c.gopDatabase, cloned)
	case 8: // 音频
		if flv.IsAACSequenceHeader(srcBytes) {
			c.aSeqHeader = cloned
			return
		}
		// 规范2：如果当前还没有累积到任何视频关键帧，不要盲目缓存孤立的音频帧
		if len(c.gopDatabase) > 0 {
			c.gopDatabase = append(c.gopDatabase, cloned)
		}
	}
}

func (c *GoCache) SendToConsumer(sendFunc func(*rtmp.Message)) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 1. 首先下发最基础的解码配置序列头(SPS/PPS以及 AAC AudioSpecifiConfig)
	if c.vSeqHeader != nil {
		// 发送时同样做一层防御性克隆，防止多用户并发读取同一Buffer 改变读写指针
		sendFunc(&rtmp.Message{
			TypeID:    c.vSeqHeader.TypeID,
			Length:    c.vSeqHeader.Length,
			Timestamp: 0, //初始头时间戳归0 //c.vSeqHeader.Timestamp,
			StreamID:  c.vSeqHeader.StreamID,
			Payload:   bytes.NewBuffer(c.vSeqHeader.Payload.Bytes()),
		})
	}

	if c.aSeqHeader != nil {
		sendFunc(&rtmp.Message{
			TypeID:    c.aSeqHeader.TypeID,
			Length:    c.aSeqHeader.Length,
			Timestamp: 0, //初始头时间戳归0
			StreamID:  c.aSeqHeader.StreamID,
			Payload:   bytes.NewBuffer(c.aSeqHeader.Payload.Bytes()),
		})
	}

	if len(c.gopDatabase) == 0 {
		return
	}

	// 2. 对当前正在发送的GOP数据库拷贝进行严格的时间戳升序重拍
	// 测地消灭 DTS out of oder 逆序错误
	sortedGop := make([]*rtmp.Message, len(c.gopDatabase))
	copy(sortedGop, c.gopDatabase)

	sort.SliceStable(sortedGop, func(i, j int) bool {
		return sortedGop[i].Timestamp < sortedGop[j].Timestamp
	})

	// 3. 按绝对递增的顺序来群发广播
	for _, msg := range sortedGop {
		sendFunc(&rtmp.Message{
			TypeID:    msg.TypeID,
			Length:    msg.Length,
			Timestamp: msg.Timestamp,
			StreamID:  msg.StreamID,
			Payload:   bytes.NewBuffer(msg.Payload.Bytes()),
		})
	}
}
