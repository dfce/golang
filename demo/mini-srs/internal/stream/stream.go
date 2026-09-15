package stream

import (
	"bytes"
	"errors"
	"fmt"
	"mini-srs/internal/protocol/rtmp"
	"sync"
)

// 对应SRS的Source或MediaMTX的Path.负责汇聚一个流的发布端，并向下广播到各个客户端Consumer

type StreamPath struct {
	path        string
	publisher   Publisher
	subscribers map[string]Subscriber
	cache       *GoCache
	mu          sync.RWMutex
}

func NewStreamPath(path string) *StreamPath {
	return &StreamPath{
		path:        path,
		subscribers: make(map[string]Subscriber),
		cache:       NewGoCache(),
	}
}

func (sp *StreamPath) SetPublisher(pub Publisher) error {
	sp.mu.Lock()
	if sp.publisher != nil {
		sp.mu.Unlock()
		err := fmt.Sprintf("stream path %s already has a publisher", sp.path)
		return errors.New(err)
	}

	sp.publisher = pub
	sp.mu.Unlock()

	go sp.transmitLoop()
	return nil
}

func (sp *StreamPath) AddSubscriber(sub Subscriber) {
	sp.mu.Lock()
	sp.subscribers[sub.GetID()] = sub
	sp.mu.Unlock()

	// 发送历史纯净GOP缓存 实现秒开
	sp.cache.SendToConsumer(func(msg *rtmp.Message) {
		_ = sub.SendMsg(msg)
	})
}
func (sp *StreamPath) RemoveSubscriber(id string) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if sub, exists := sp.subscribers[id]; exists {
		sub.Close()
		delete(sp.subscribers, id)
	}
}
func (sp *StreamPath) transmitLoop() {
	// 收集需要被移除的死连接，避免在map循环迭代时直接删除
	var deadSubscribers []string
	defer func() {
		sp.mu.Lock()
		for id, sub := range sp.subscribers {
			sub.Close()
			if sess, ok := sub.(interface{ NotifyDone() }); ok {
				sess.NotifyDone()
			}
			delete(sp.subscribers, id)
		}

		if sp.publisher != nil {
			sp.publisher.NotifyDone()
		}

		sp.publisher = nil
		sp.mu.Unlock()
		// 【核心修复】：推流彻底断开后，安全的从全局管理器中把自己移除
		DefaultManager.RemovePath(sp.path)
		fmt.Printf("[Mini-SRS] 通道 [%s] 没有任何活动， 已安全销毁\n", sp.path)
	}()

	for {
		msg, err := sp.publisher.ReadMsg()
		if err != nil {
			break
		}
		// 1. 将数据喂入GOP缓存管理器(内部已做深拷贝)
		sp.cache.Cache(msg)

		// 2.	针对所有当前在线观众，必须将实时帧进行一份临时的、完全独立的深拷贝
		// 这样彻底断绝了推流主协程在下一微妙调用 ReadMsg() 对发送队列数据的篡改
		srcBytes := msg.Payload.Bytes()
		if len(srcBytes) == 0 {
			continue
		}

		sp.mu.RLock()
		for id, sub := range sp.subscribers {
			// 为每一个在线观众克隆独立的物理内存切片
			liveBytes := make([]byte, len(srcBytes))
			copy(liveBytes, srcBytes)

			liveMsg := &rtmp.Message{
				TypeID:    msg.TypeID,
				Length:    msg.Length,
				Timestamp: msg.Timestamp,
				StreamID:  msg.StreamID,
				Payload:   bytes.NewBuffer(liveBytes),
			}

			if err := sub.SendMsg(liveMsg); err != nil {
				deadSubscribers = append(deadSubscribers, id)
			}
		}
		sp.mu.RUnlock()

		// 3. 集中安全清理：如果发现了死连接，退出读锁后，统一加写锁剔除
		if len(deadSubscribers) > 0 {
			sp.mu.Lock()
			for _, id := range deadSubscribers {
				if sub, exists := sp.subscribers[id]; exists {
					sub.Close()

					// 尝试反向将Session的doneCh 唤醒
					if sess, ok := sub.(interface{ NotifyDone() }); ok {
						sess.NotifyDone()
					}

					delete(sp.subscribers, id)
					fmt.Printf("[Mini-SRS] 观众因网络错误自动移出通道：%s\n", id)
				}
			}
			sp.mu.Unlock()
			deadSubscribers = deadSubscribers[:0] // 清空暂存切片复用
		}
	}
}
