package messages

import "bytes"

type Type int

const (
	TypeUnknown Type = iota
	TypeHello
	TypeRegister
	TypeNotification
)

var typeField = []byte(`"messageType":"`)

func MessageType(msg []byte) Type {
	idx := bytes.Index(msg, typeField)
	if idx == -1 {
		return TypeUnknown
	}
	start := idx + len(typeField)

	idx = bytes.IndexByte(msg[start:], '"')
	if idx == -1 {
		return TypeUnknown
	}
	end := start + idx

	switch string(msg[start:end]) {
	case "register":
		return TypeRegister
	case "hello":
		return TypeHello
	case "notification":
		return TypeNotification
	}

	return TypeUnknown
}
