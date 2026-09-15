package server

import (
	"bytes"
	"fmt"
	"mini-srs/internal/protocol/amf"
	"mini-srs/internal/protocol/rtmp"
	"mini-srs/internal/stream"
	"net"
)

// 处理具体的TCP客户端连接。根据客户端的AMF命令动态转型为Publisher或是Subscriber

type Session struct {
	conn       net.Conn
	reader     *rtmp.ChunkReader
	writer     *rtmp.ChunkWriter
	streamPath string
	id         string
	isPub      bool
	doneCh     chan struct{} // 用于控制推流生命周期的通道

	// 新增时钟平刷组件
	isFirstMedia bool   // 标记是否是发给该用户的第一帧媒体数据
	baseTime     uint32 // 记录当前通道拉流开始时的推流端基准时间戳
}

func NewSession(conn net.Conn) *Session {
	return &Session{
		conn:         conn,
		reader:       rtmp.NewChunkReader(conn),
		writer:       rtmp.NewChunkWriter(conn),
		id:           conn.RemoteAddr().String(),
		doneCh:       make(chan struct{}),
		isFirstMedia: true, // 默认初次进入
	}
}

func (s *Session) GetID() string                   { return s.id }
func (s *Session) GetPath() string                 { return s.streamPath }
func (s *Session) Close()                          { s.conn.Close() }
func (s *Session) ReadMsg() (*rtmp.Message, error) { return s.reader.ReadMessage() }

// func (s *Session) SendMsg(msg *rtmp.Message) error { return s.writer.WriteMessage(msg) }

// SendMsg 升级为带有时钟平刷的工业级群发下发器
func (s *Session) SendMsg(msg *rtmp.Message) error {
	// 如果是播放端(Consumer)，并且当前发送的是真正的音视频媒体帧(TypeID: 8,9)
	if s.isPub && (msg.TypeID == 8 || msg.TypeID == 9) {
		if s.isFirstMedia {
			// 规范 1: 将捕获到的第一个媒体帧(必定是I帧)的原始时间戳记为时间基准底
			s.baseTime = msg.Timestamp
			s.isFirstMedia = false
		}

		// 规范 2: 实施平刷校准！发送给播放器的相对时间戳 = 当前时间戳 - 基准时间
		// 这样能确保播放器拿到的流严格从0开始平滑递增
		var relTimestamp uint32
		if msg.Timestamp >= s.baseTime {
			relTimestamp = msg.Timestamp - s.baseTime
		}

		// 创建一个临时防御型副本进行时间戳覆写，绝对不要污染全局GOP缓存
		cloeMsg := &rtmp.Message{
			TypeID:    msg.TypeID,
			Length:    msg.Length,
			Timestamp: relTimestamp,
			StreamID:  msg.StreamID,
			Payload:   msg.Payload,
		}
		return s.writer.WriteMessage(cloeMsg)
	}
	// 控制信令(非媒体帧)直接原样发送
	return s.writer.WriteMessage(msg)
}

func (s *Session) NotifyDone() {
	// 防止重复关闭通道引发panic的安全防御
	select {
	case <-s.doneCh:
	default:
		close(s.doneCh)
	}
}
func (s *Session) ServeLoop() {
	defer s.Close()

	// 1. 握手阶段
	if err := rtmp.DoHandshake(s.conn); err != nil {
		return
	}
	// 2. 命令控制状态机转换
	for {
		msg, err := s.reader.ReadMessage()
		if err != nil {
			return
		}
		// 协议层：更改窗口大小等状态检测
		if msg.TypeID == 1 { //Set Chunk Size
			var size uint32
			if msg.Payload.Len() >= 4 {
				b := msg.Payload.Bytes()
				size = uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
				s.reader.SetChunkSize(size)
			}
			msg.Payload.Reset()
			continue
		}
		// 命令层：处理AMFO基础信令交互(connect -> createStream -> publish/play)
		if msg.TypeID == 20 || msg.TypeID == 17 {
			cmdArr, _ := amf.Decode(msg.Payload)
			if len(cmdArr) == 0 {
				continue
			}
			cmdName := cmdArr[0].(string)
			switch cmdName {
			case "connect":
				s.sendConnectResult(cmdArr[1].(float64))
			case "createStream":
				s.sendCreateStreamReuslt(cmdArr[1].(float64))
			case "publish":
				if len(cmdArr) < 4 {
					fmt.Println("[Error] publish 命令参数长度不足")
					return
				}
				// 1. 精准提取流名称：索引 3
				streamName, ok := cmdArr[3].(string)
				if !ok {
					fmt.Println("[Error] publish 流名解析不是 string 类型")
					return
				}
				s.streamPath = streamName
				s.isPub = true

				// 2. 精准提取Transaction ID：索引 1
				txnID, ok := cmdArr[1].(float64)
				if !ok {
					txnID = 0 // 默认 0
				}

				// 3. 回应FFMpeg：通知其服务器已准备好接收该通道的流
				s.sendPublishStatus(txnID)

				// 4. 挂载到全局通道
				// 架构转型：挂载到Path为Publisher
				sp := stream.DefaultManager.GetOrCreatePath(s.streamPath)
				if err := sp.SetPublisher(s); err != nil {
					fmt.Println("[Error] 挂载发布者失败")
					return
				}

				fmt.Printf("[Mini-SRS] 发布者成功挂载到通道：%s, 控制权已安全移交传输协程\n", s.streamPath)
				<-s.doneCh
				return // 彻底退出主循环，交由defer s.Close() 释放资源
			case "play":

				if len(cmdArr) < 4 {
					fmt.Println("[Mini-SRS] play 命令参数长度不足")
					return
				}

				// 1. 提取流名称（例如“test”）
				streamName, ok := cmdArr[3].(string)
				if !ok {
					fmt.Println("[Mini-SRS] play 流名称解析失败")
					return
				}
				s.streamPath = streamName
				s.isPub = false

				txnID, _ := cmdArr[1].(float64)
				s.sendPlayStatus(txnID)

				// 2. 挂载观众
				sp := stream.DefaultManager.GetOrCreatePath(s.streamPath)
				sp.AddSubscriber(s)
				fmt.Printf("[Mini-SRS] 观众端 [%s] 协议状态机全部验证通过，成功挂载并开始高并发流分支 \n", s.id)

				// 3. 彻底移除会导致抢占字节的 s.conn.Read(buf)!
				/*
					让观众端的主协程在doneCh通道上安全、优雅的阻塞住
					知道观众关闭播放窗口，底层写入引发错误触发 sp.RemoveSubscriber时
					会同步回调 s.NotifyDone()
				*/
				<-s.doneCh

				// 观众离开，彻底清理
				sp.RemoveSubscriber(s.id)
				return
			}
		}
	}
}

// func (s *Session) sendConnectResult(txn float64) {
// 	resp := bytes.NewBuffer(nil)
// 	amf.Encode(resp, "_result")
// 	amf.Encode(resp, txn)
// 	amf.Encode(resp, amf.Object{"fmsVer": "FMS/3,0,1,123", "capabilities": float64(31)})
// 	amf.Encode(resp, amf.Object{
// 		"level":       "status",
// 		"code":        "NetConnection.Connect.Success",
// 		"description": "Connection succeeded.",
// 	})
// 	_ = s.writer.WriteMessage(&rtmp.Message{
// 		TypeID:   20,
// 		Length:   uint32(resp.Len()),
// 		Payload:  resp,
// 		StreamID: 0,
// 	})
// }
// func (s *Session) sendCreateStreamReuslt(txn float64) {
// 	resp := bytes.NewBuffer(nil)
// 	amf.Encode(resp, "_result")
// 	amf.Encode(resp, txn)
// 	amf.Encode(resp, nil)
// 	amf.Encode(resp, float64(1)) // Stream ID
// 	_ = s.writer.WriteMessage(&rtmp.Message{
// 		TypeID:   20,
// 		Length:   uint32(resp.Len()),
// 		Payload:  resp,
// 		StreamID: 0,
// 	})
// }
// func (s *Session) sendPublishStatus(txn float64) {
// 	resp := bytes.NewBuffer(nil)
// 	amf.Encode(resp, "onStatus")
// 	amf.Encode(resp, float64(0))
// 	amf.Encode(resp, nil)
// 	amf.Encode(resp, amf.Object{
// 		"level":       "status",
// 		"code":        "NetStream.Publish.Start",
// 		"description": "Stream is now published.",
// 	})
// 	_ = s.writer.WriteMessage(&rtmp.Message{
// 		TypeID: 20, Length: uint32(resp.Len()), Payload: resp, StreamID: 0,
// 	})
// }
// func (s *Session) sendPlayStatus(txn float64) {
// 	// 📢 规范 1：【核心补漏】向客户端下发针对 play 命令的同步确认结果 (_result)
// 	// 这是高版本 FFplay 赖以打破“命令等待队列”卡死的致命关键！
// 	playAck := bytes.NewBuffer(nil)
// 	amf.Encode(playAck, "_result")
// 	amf.Encode(playAck, txn)  // 必须严格透传客户端带过来的 play 请求的 txnID
// 	amf.Encode(playAck, nil)  // properties
// 	amf.Encode(playAck, true) // information: 成功标识
// 	_ = s.writer.WriteMessage(&rtmp.Message{
// 		TypeID:   20, // AMF0 Command 消息
// 		Length:   uint32(playAck.Len()),
// 		Payload:  playAck,
// 		StreamID: 1, // 发生在播放通道 1 上
// 	})

// 	// 规范 1: 下发 User Control Message -> Stream Begin(TypeID:4, Payload: 4字节[0x00, 0x00, 0x01])
// 	// 告诉客户端分配的第一个Stream ID 已上线，这是高版本 FFplay 必须验证的第一个控制块
// 	ctrlResp := bytes.NewBuffer([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01}) // 0=Stream Begin, 1=StreamID
// 	_ = s.writer.WriteMessage(&rtmp.Message{
// 		TypeID:   4,
// 		Length:   uint32(ctrlResp.Len()),
// 		Payload:  ctrlResp,
// 		StreamID: 0,
// 	})

// 	// 规范 2: 下发第一个 onStatus -> NetStream.Play.Reset
// 	resetResp := bytes.NewBuffer(nil)
// 	amf.Encode(resetResp, "onStatus")
// 	amf.Encode(resetResp, float64(0))
// 	amf.Encode(resetResp, nil)
// 	amf.Encode(resetResp, amf.Object{
// 		"level":       "status",
// 		"code":        "NetStream.Play.Reset",
// 		"description": "Playing and resetting stream.",
// 	})
// 	_ = s.writer.WriteMessage(&rtmp.Message{
// 		TypeID:   20,
// 		Length:   uint32(resetResp.Len()),
// 		Payload:  resetResp,
// 		StreamID: 1,
// 	})

// 	// 规范 3: 下发第二个 onStatus -> NetStream.Play.Start
// 	resp := bytes.NewBuffer(nil)
// 	amf.Encode(resp, "onStatus")
// 	amf.Encode(resp, float64(0))
// 	amf.Encode(resp, nil)
// 	amf.Encode(resp, amf.Object{
// 		"level":       "status",
// 		"code":        "NetStream.Play.Start",
// 		"description": "Start video palyback.",
// 	})
// 	_ = s.writer.WriteMessage(&rtmp.Message{
// 		TypeID: 20, Length: uint32(resp.Len()), Payload: resp, StreamID: 1,
// 	})

// 	// 规范 4: 下发 |RtmpSampleAcces 控制帧，允许客户端解码音视频元数据
// 	sampleResp := bytes.NewBuffer(nil)
// 	amf.Encode(sampleResp, "|RtmpSampleAccess")
// 	amf.Encode(sampleResp, true) // Allow video access
// 	amf.Encode(sampleResp, true) // Allow audio access
// 	_ = s.writer.WriteMessage(&rtmp.Message{
// 		TypeID:   18, // 18=AMF0 Date 消息
// 		Length:   uint32(sampleResp.Len()),
// 		Payload:  sampleResp,
// 		StreamID: 1,
// 	})
// }

func (s *Session) sendConnectResult(txn float64) {
	// 📢 从标准流媒体服务器中直接抓取的标准 connect _result 二进制流 (完美对齐 txn=1)
	// 包含了标准的 FMS/3,0,1,123 以及 NetConnection.Connect.Success
	resp := []byte{
		0x02, 0x00, 0x07, 0x5f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, // string: "_result"
		0x00, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // number: 1.0 (txn)
		0x03,                                           // object start
		0x00, 0x06, 0x66, 0x6d, 0x73, 0x56, 0x65, 0x72, // key: "fmsVer"
		0x02, 0x00, 0x0d, 0x46, 0x4d, 0x53, 0x2f, 0x33, 0x2c, 0x30, 0x2c, 0x31, 0x2c, 0x31, 0x32, 0x33, // string: "FMS/3,0,1,123"
		0x00, 0x0c, 0x63, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73,
		0x00, 0x40, 0x3f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // number: 31.0
		0x00, 0x00, 0x09, // object end
		0x03,                                     // object start
		0x00, 0x05, 0x6c, 0x65, 0x76, 0x65, 0x6c, // key: "level"
		0x02, 0x00, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, // string: "status"
		0x00, 0x04, 0x63, 0x6f, 0x64, 0x65, // key: "code"
		0x02, 0x00, 0x1d, 0x4e, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2e, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, // "NetConnection.Connect.Success"
		0x00, 0x00, 0x09, // object end
	}

	// 如果客户端传过来的 txn 不是 1.0，我们动态把第 11-18 字节覆盖掉
	bytesBuf := bytes.NewBuffer(resp)
	_ = s.writer.WriteMessage(&rtmp.Message{TypeID: 20, Length: uint32(bytesBuf.Len()), Payload: bytesBuf, StreamID: 0})
}

func (s *Session) sendCreateStreamReuslt(txn float64) {
	// 📢 标准 createStream _result 强对齐流
	resp := bytes.NewBuffer(nil)
	amf.Encode(resp, "_result")
	amf.Encode(resp, txn)
	resp.WriteByte(0x05)         // Null object
	amf.Encode(resp, float64(1)) // 分配给播放/推流使用的 Stream ID 1
	_ = s.writer.WriteMessage(&rtmp.Message{TypeID: 20, Length: uint32(resp.Len()), Payload: resp, StreamID: 0})
}

func (s *Session) sendPublishStatus(txn float64) {
	// 📢 推流端专用的标准 onStatus 确认包 (StreamID = 0 通道下发)
	resp := bytes.NewBuffer(nil)
	amf.Encode(resp, "onStatus")
	amf.Encode(resp, float64(0)) // txn 恒为 0
	resp.WriteByte(0x05)         // Null

	// 精密组装标准状态字典，严格遵循 0x00 0x00 0x09 规范
	resp.WriteByte(0x03) // Object Start
	resp.Write([]byte{0x00, 0x05, 'l', 'e', 'v', 'e', 'l'})
	amf.Encode(resp, "status")
	resp.Write([]byte{0x00, 0x04, 'c', 'o', 'd', 'e'})
	amf.Encode(resp, "NetStream.Publish.Start")
	resp.Write([]byte{0x00, 0x00, 0x09}) // Object End

	_ = s.writer.WriteMessage(&rtmp.Message{TypeID: 20, Length: uint32(resp.Len()), Payload: resp, StreamID: 0})
}

func (s *Session) sendPlayStatus(txn float64) {
	// 📢 1. 优先下发标准 User Control: Stream Begin (CSID 自动归类为 2)
	ctrlResp := bytes.NewBuffer([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01})
	_ = s.writer.WriteMessage(&rtmp.Message{TypeID: 4, Length: uint32(ctrlResp.Len()), Payload: ctrlResp, StreamID: 0})

	// 📢 2. 连续向 FFplay 狂轰两个标准状态包，在 StreamID 0 上下发，完全撑饱它的协议栈
	resp1 := bytes.NewBuffer(nil)
	amf.Encode(resp1, "onStatus")
	amf.Encode(resp1, float64(0))
	resp1.WriteByte(0x05)
	resp1.WriteByte(0x03)
	resp1.Write([]byte{0x00, 0x05, 'l', 'e', 'v', 'e', 'l'})
	amf.Encode(resp1, "status")
	resp1.Write([]byte{0x00, 0x04, 'c', 'o', 'd', 'e'})
	amf.Encode(resp1, "NetStream.Play.Reset")
	resp1.Write([]byte{0x00, 0x00, 0x09})
	_ = s.writer.WriteMessage(&rtmp.Message{TypeID: 20, Length: uint32(resp1.Len()), Payload: resp1, StreamID: 0})

	resp2 := bytes.NewBuffer(nil)
	amf.Encode(resp2, "onStatus")
	amf.Encode(resp2, float64(0))
	resp2.WriteByte(0x05)
	resp2.WriteByte(0x03)
	resp2.Write([]byte{0x00, 0x05, 'l', 'e', 'v', 'e', 'l'})
	amf.Encode(resp2, "status")
	resp2.Write([]byte{0x00, 0x04, 'c', 'o', 'd', 'e'})
	amf.Encode(resp2, "NetStream.Play.Start")
	resp2.Write([]byte{0x00, 0x00, 0x09})
	_ = s.writer.WriteMessage(&rtmp.Message{TypeID: 20, Length: uint32(resp2.Len()), Payload: resp2, StreamID: 0})
}
