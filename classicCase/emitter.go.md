# go 实现推送消息 与 node socket.io 通讯

```go
package emitter

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

type packetType int

const (
	TypeConnect packetType = iota
	TypeDisconnect
	TypeEvent
	TypeAck
	TYpeConnectError
)

type Packet struct {
	Type packetType  `msgpack:"type"`
	Data interface{} `msgpack:"data"`
	Nsp  string      `msgpack:"nsp"`
}

type Options struct {
	Rooms  []string               `msgpack:"rooms,omitempty"`
	Execpt []string               `msgpack:"except,omitempty"`
	Flags  map[string]interface{} `msgpack:"flags,omitempty"`
}

type Emitter struct {
	rdb     *redis.Client
	nsp     string
	rooms   []string
	except  []string
	flags   map[string]interface{}
	channel string
}

func NewEmitter(rdb *redis.Client) *Emitter {
	return &Emitter{
		rdb:     rdb,
		nsp:     "/",
		channel: "socket.io",
		rooms:   []string{},
		except:  []string{},
		flags:   map[string]interface{}{},
	}
}

func (e *Emitter) Of(nsp string) *Emitter {
	cp := *e
	cp.nsp = nsp
	return &cp
}

func (e *Emitter) To(room string) *Emitter {
	cp := *e
	cp.rooms = append(cp.rooms, room)
	return &cp
}

func (e *Emitter) In(room string) *Emitter {
	return e.To(room)
}

// except rooms
func (e *Emitter) Except(room string) *Emitter {
	cp := *e
	cp.except = append(cp.except, room)
	return &cp
}

func (e *Emitter) WithFlag(key string, value interface{}) *Emitter {
	cp := *e
	cp.flags[key] = value
	return &cp
}

func (e *Emitter) Emit(event string, data interface{}) error {
	packet := Packet{
		Type: TypeEvent,
		Data: []interface{}{event, data},
		Nsp:  e.nsp,
	}

	uid := randomUID()

	opts := Options{
		Rooms:  e.rooms,
		Execpt: e.except,
		Flags:  e.flags,
	}

	msg := []interface{}{uid, packet, opts}
	fmt.Printf("msg: %+v", msg)
	buf, err := msgpack.Marshal(msg)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf("%s#%s#", e.channel, e.nsp)
	ctx := context.Background()
	fmt.Println("channel:", channel)
	return e.rdb.Publish(ctx, channel, buf).Err()
}

func randomUID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

```