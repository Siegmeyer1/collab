package messages

import (
	"bytes"
	"encoding/binary"
)

type Protocol uint64

var (
	SyncProtocol      Protocol = 0
	AwarenessProtocol Protocol = 1
)

type MessageType uint64

var (
	SyncStep1 MessageType = 0
	SyncStep2 MessageType = 1
	Update    MessageType = 2
)

type Message struct {
	Protocol    Protocol
	MessageType MessageType // only valid for Protocol=SyncProtocol
	Data        []byte
}

func DecodeMessage(b []byte) (*Message, error) {
	buf := bytes.NewBuffer(b)
	var protocol Protocol
	var messageType MessageType

	p, err := binary.ReadUvarint(buf)
	if err != nil {
		return nil, err
	}
	protocol = Protocol(p)

	if protocol == SyncProtocol {
		mt, err := binary.ReadUvarint(buf)
		if err != nil {
			return nil, err
		}
		messageType = MessageType(mt)
	}

	return &Message{
		Protocol:    protocol,
		MessageType: messageType,
		Data:        b,
	}, nil
}
