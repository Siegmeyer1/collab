package messages

import (
	"bytes"
	"encoding/binary"
)

func readStateVectorElement(buf *bytes.Buffer) (StateVectorElement, error) {
	clientID, err := binary.ReadUvarint(buf)
	if err != nil {
		return StateVectorElement{}, err
	}

	clock, err := binary.ReadUvarint(buf)
	if err != nil {
		return StateVectorElement{}, err
	}

	return StateVectorElement{ClientID: clientID, Clock: clock}, nil
}

var skip = uint8(10)
var item = uint8(31)

func decodeStruct(buf *bytes.Buffer) error {
	var header uint8

	if err := binary.Read(buf, binary.LittleEndian, &header); err != nil {
		return err
	}

	if header == skip {
		if _, err := binary.ReadUvarint(buf); err != nil {
			return err
		}
		return nil
	}

	if header&item != 0 {
		if err := readMeta(buf, header); err != nil {
			return err
		}

	}

	return nil
}

var leftID = uint8(1 << 7)
var rightID = uint8(1 << 6)
var attr = uint8(1 << 5)

func readMeta(buf *bytes.Buffer, header uint8) error {
	if header&leftID != 0 {
		if _, err := readStateVectorElement(buf); err != nil {
			return err
		}
	}

	if header&rightID != 0 {
		if _, err := readStateVectorElement(buf); err != nil {
			return err
		}
	}

	if header&(leftID|rightID) == 0 {
		info, err := binary.ReadUvarint(buf)
		if err != nil {
			return err
		}

		if info == 1 {
			if _, err := readArray(buf); err != nil {
				return err
			}
		} else {
			if _, err := readStateVectorElement(buf); err != nil {
				return err
			}
		}

		if header&attr != 0 {
			if _, err := readArray(buf); err != nil {
				return err
			}
		}
	}

	return nil
}

//func readItem(buf *bytes.Buffer, header uint8) ([]byte, error) {
//	itemType := header & item
//
//	switch itemType {
//
//	}
//}

func readArray(buf *bytes.Buffer) ([]byte, error) {
	l, err := binary.ReadUvarint(buf)
	if err != nil {
		return nil, err
	}

	b := make([]byte, l)
	err = binary.Read(buf, binary.LittleEndian, b)

	if err != nil {
		return nil, err
	}

	return b, nil
}
