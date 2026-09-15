package amf

import (
	"bytes"
	"encoding/binary"
)

func Encode(w *bytes.Buffer, val any) {
	if val == nil {
		w.WriteByte(TypeNull)
		return
	}
	switch v := val.(type) {
	case float64:
		w.WriteByte(TypeNumber)
		_ = binary.Write(w, binary.BigEndian, v)
	case string:
		w.WriteByte(TypeString)
		_ = binary.Write(w, binary.BigEndian, uint16(len(v)))
		w.WriteString(v)
	case bool:
		w.WriteByte(TypeBoolean)
		if v {
			w.WriteByte(1)
		} else {
			w.WriteByte(0)
		}
	case Object:
		w.WriteByte(TypeObject)
		for k, valItem := range v {
			_ = binary.Write(w, binary.BigEndian, uint16(len(k)))
			w.WriteString(k)
			Encode(w, valItem)
		}
		_ = binary.Write(w, binary.BigEndian, uint16(0))
		w.WriteByte(TypeObjectEnd)
	}
}
