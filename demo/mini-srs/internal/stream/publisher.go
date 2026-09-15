package stream

import "mini-srs/internal/protocol/rtmp"

type Publisher interface {
	GetPath() string
	ReadMsg() (*rtmp.Message, error)
	NotifyDone()
	Close()
}
