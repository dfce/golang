package amf

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	TypeNumber    = 0x00
	TypeBoolean   = 0x01
	TypeString    = 0x02
	TypeObject    = 0x03
	TypeNull      = 0x05
	TypeECMArray  = 0x08
	TypeObjectEnd = 0x09
)

type Object map[string]any // map[string]interface{}

func Decode(r io.Reader) ([]any, error) {
	var result []any
	for {
		val, err := DecoderOne(r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		result = append(result, val)
	}
	return result, nil
}

func DecoderOne(r io.Reader) (any, error) {
	var marker uint8
	if err := binary.Read(r, binary.BigEndian, &marker); err != nil {
		return nil, err
	}

	switch marker {
	case TypeNumber:
		var val float64
		if err := binary.Read(r, binary.BigEndian, &val); err != nil {
			return nil, err
		}
		return val, nil
	case TypeBoolean:
		var val uint8
		if err := binary.Read(r, binary.BigEndian, &val); err != nil {
			return nil, err
		}
		return val != 0, nil
	case TypeString:
		return readString(r)
	case TypeObject:
		obj := make(Object)
		for {
			key, err := readString(r)
			if err != nil {
				return nil, err
			}
			var endMarker uint8
			if key == "" {
				if err := binary.Read(r, binary.BigEndian, &endMarker); err == nil && endMarker == TypeObjectEnd {
					break
				}
			}
			val, err := DecoderOne(r)
			if err != nil {
				return nil, err
			}
			obj[key] = val
		}
		return obj, nil

	case TypeNull:
		return nil, nil
	case TypeECMArray:
		var size uint32
		binary.Read(r, binary.BigEndian, &size)
		obj := make(Object)
		for i := uint32(0); i < size; i++ {
			key, err := readString(r)
			if err != nil {
				return nil, err
			}
			val, err := DecoderOne(r)
			if err != nil {
				return nil, err
			}
			obj[key] = val
		}
		var endMarker [3]byte
		_, _ = r.Read(endMarker[:]) // Skip 0x00 0x00 0x09
		return obj, nil
	default:
		return nil, errors.New("ubknow amf0 type")
	}
}

func readString(r io.Reader) (string, error) {
	var length uint16
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return "", err
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}

	return string(buf), nil
}
