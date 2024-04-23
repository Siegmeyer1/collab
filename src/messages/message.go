package messages

import (
	"bytes"
	"diploma/src/logging"
	"encoding/binary"
	"fmt"
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

func PeekProtoAndType(b []byte) (Protocol, MessageType, error) {
	var protocol Protocol
	var messageType MessageType

	p, n := binary.Uvarint(b[0:1])
	if n <= 0 {
		return 0, 0, fmt.Errorf("bad protocol byte, read %d bytes", n)
	}
	protocol = Protocol(p)

	if protocol == SyncProtocol {
		t, n := binary.Uvarint(b[1:2])
		if n <= 0 {
			return 0, 0, fmt.Errorf("bad type byte, read %d bytes", n)
		}
		messageType = MessageType(t)
	}

	return protocol, messageType, nil
}

func ReadProtoAndType(buffer *bytes.Buffer) (Protocol, MessageType, error) {
	var protocol Protocol
	var messageType MessageType

	p, err := binary.ReadUvarint(buffer)
	if err != nil {
		return 0, 0, err
	}
	protocol = Protocol(p)

	if protocol == SyncProtocol {
		mt, err := binary.ReadUvarint(buffer)
		if err != nil {
			return 0, 0, err
		}
		messageType = MessageType(mt)
	}

	return protocol, messageType, nil
}

func DecodeMessage(b []byte) (*Message, error) {
	buf := bytes.NewBuffer(b)

	protocol, messageType, err := ReadProtoAndType(buf)
	if err != nil {
		return nil, err
	}

	return &Message{
		Protocol:    protocol,
		MessageType: messageType,
		Data:        b,
	}, nil
}

type StateVectorElement struct {
	ClientID uint64
	Clock    uint64
}

type Step1SyncMessage struct {
	StateVector []StateVectorElement
}

func DecodeStep1SyncMessage(b []byte) (*Step1SyncMessage, error) {
	buf := bytes.NewBuffer(b)

	protocol, messageType, err := ReadProtoAndType(buf)
	if err != nil {
		return nil, err
	}

	if protocol != SyncProtocol {
		return nil, fmt.Errorf("decoding Step1Sync msg: wrong protocol: %d", protocol)
	}

	if messageType != SyncStep1 {
		return nil, fmt.Errorf("decoding Step1Sync msg: wrong messageType: %d", messageType)
	}

	msg := &Step1SyncMessage{}

	_, err = binary.ReadUvarint(buf) // this is num of bytes left in message, we don`t need this
	if err != nil {
		return nil, err
	}

	svLength, err := binary.ReadUvarint(buf)
	if err != nil {
		return nil, err
	}

	for i := uint64(0); i < svLength; i++ {
		el, err := readStateVectorElement(buf)
		if err != nil {
			return nil, fmt.Errorf("reading element [%d]: %w", i, err)
		}

		msg.StateVector = append(msg.StateVector, el)
	}

	return msg, nil
}

type UpdateMessage struct {
	IsDeleteOnly bool
	ClientID     uint64
	Clock        uint64
	Data         []byte
	DeleteData   []byte // only valid if IsDeleteOnly is true
}

func DecodeUpdateMessage(b []byte) (*UpdateMessage, error) {
	buf := bytes.NewBuffer(b)
	var isDeleteOnly bool
	var clientID, clock uint64
	var deleteData []byte

	protocol, messageType, err := ReadProtoAndType(buf)
	if err != nil {
		return nil, err
	}

	if protocol != SyncProtocol {
		return nil, fmt.Errorf("decoding update msg: wrong protocol: %d", protocol)
	}

	if messageType != Update {
		return nil, fmt.Errorf("decoding update msg: wrong messageType: %d", messageType)
	}

	_, err = binary.ReadUvarint(buf) //  bytes left?
	if err != nil {
		return nil, err
	}

	numOfUpdates, err := binary.ReadUvarint(buf)
	if err != nil {
		return nil, err
	}
	if numOfUpdates > 1 {
		return nil, fmt.Errorf("unexpected num of updates: %d", numOfUpdates)
	}

	if numOfUpdates == 1 {
		isDeleteOnly = false

		numOfStructs, err := binary.ReadUvarint(buf) // this is number of structs in update, omit it for now
		if err != nil {
			return nil, err
		}
		logging.Debug("numOfStructs: %d", numOfStructs)

		clientID, err = binary.ReadUvarint(buf)
		if err != nil {
			return nil, err
		}
		clock, err = binary.ReadUvarint(buf)
		if err != nil {
			return nil, err
		}

	} else {
		isDeleteOnly = true

		deleteData = buf.Bytes()
	}

	msg := &UpdateMessage{
		IsDeleteOnly: isDeleteOnly,
		ClientID:     clientID,
		Clock:        clock,
		Data:         b,
		DeleteData:   deleteData,
	}

	return msg, nil
}

//func DecodeDeleteData(b []byte) (map[uint64]uint64, error) {
//	buf := bytes.NewBuffer(b)
//
//	numOfClients, err := binary.ReadUvarint(buf)
//	if err != nil {
//		return nil, err
//	}
//
//	for i := uint64(0); i < numOfClients; i++ {
//		clientID, err := binary.ReadUvarint(buf)
//		if err != nil {
//			return nil, err
//		}
//
//		numOfDeletes, err := binary.ReadUvarint(buf)
//		if err != nil {
//			return nil, err
//		}
//
//		for i := uint64(0); i < numOfDeletes; i++ {
//
//		}
//	}
//}
