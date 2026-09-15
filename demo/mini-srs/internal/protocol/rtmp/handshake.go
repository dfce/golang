package rtmp

/*
 手动实现标准的RTMP简单握手协议（C0S0，C1S1，C2S2状态机）
*/

import (
	"crypto/rand"
	"errors"
	"io"
	"net"
)

func DoHandshake(conn net.Conn) error {
	// 1. Read C0 + C1
	c0c1 := make([]byte, 1+1536)
	if _, err := io.ReadFull(conn, c0c1); err != nil {
		return err
	}

	if c0c1[0] != 3 {
		return errors.New("only support RTMP varsion 3")
	}
	// 2. 构造服务器响应：S0(1字节) + S1(1536字节) + S2(1536字节)
	// 总大小： 1+1536+1536=3073
	s0s1s2 := make([]byte, 1+1536+1536)

	// S0：RTMP 版本号固定为 3
	s0s1s2[0] = 3

	// --- 构造 S1(偏移量1-1536) ---
	// A. 时间戳(4字节，设为0)
	s0s1s2[1] = 0
	s0s1s2[2] = 0
	s0s1s2[3] = 0
	s0s1s2[4] = 0

	// B. 根据标准规范，必须包含4字节的零占位符(Zero)
	s0s1s2[5] = 0
	s0s1s2[6] = 0
	s0s1s2[7] = 0
	s0s1s2[8] = 0
	// C.	随机数(1528字节)
	if _, err := rand.Read(s0s1s2[9:1537]); err != nil {
		return err
	}
	// --- 构造 S2(偏移量1537-3073) ---
	// 标准要求S2必须是客户端C1的严格镜像回显(Echo C1)
	copy(s0s1s2[1537:], c0c1[1:1537])

	// 3. 一次性将S0+S1+S2写入网络套接字
	if _, err := conn.Write(s0s1s2); err != nil {
		return err
	}
	// 4. 读取客户端发来的C2(1536字节)确认块
	c2 := make([]byte, 1536)
	if _, err := io.ReadFull(conn, c2); err != nil {
		return err
	}

	return nil
}
