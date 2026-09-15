package stream

import "mini-srs/internal/protocol/rtmp"

type Subscriber interface {
	GetID() string
	SendMsg(*rtmp.Message) error
	Close()
}
