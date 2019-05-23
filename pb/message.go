package rp

import (
	"encoding/binary"
	"errors"

	proto "github.com/golang/protobuf/proto"
	//timestamp "github.com/golang/protobuf/ptypes/timestamp"
)

type wireMessage struct {
	*Header
	*Key
	*Command
	*Terminal
	MessageType string
}

var messageMissing error = errors.New("Message is missing")
var functionMismatch error = errors.New("Function did not match message type")

func (w wireMessage) Do(F interface{}) error {
	switch w.MessageType {
	case "header":
		if w.Header == nil {
			return messageMissing
		}
		if callback, ok := F.(func(*Header) error); ok {
			return callback(w.Header)
		}
	case "key":
		if w.Key == nil {
			return messageMissing
		}
		if callback, ok := F.(func(*Key) error); ok {
			return callback(w.Key)
		}
	case "command":
		if w.Command == nil {
			return messageMissing
		}
		if callback, ok := F.(func(*Command) error); ok {
			return callback(w.Command)
		}
	case "terminal":
		if w.Terminal == nil {
			return messageMissing
		}
		if callback, ok := F.(func(*Terminal) error); ok {
			return callback(w.Terminal)
		}
	default:
		return errors.New("WireMessage unrecognized")
	}
	return functionMismatch
}

func readMessage(b []byte) (*wireMessage, error) {
	var message RecordMessage
	if err := proto.Unmarshal(b, &message); err != nil {
		return nil, err
	} else if header := message.GetHeader(); header != nil {
		return &wireMessage{Header: header, MessageType: "header"}, nil
	} else if key := message.GetKey(); key != nil {
		return &wireMessage{Key: key, MessageType: "key"}, nil
	} else if command := message.GetCommand(); command != nil {
		return &wireMessage{Command: command, MessageType: "command"}, nil
	} else if terminal := message.GetTerminal(); command != nil {
		return &wireMessage{Terminal: terminal, MessageType: "terminal"}, nil
	}
	return nil, errors.New("Unknown message type")
}

func createMessage(NM interface{}) ([]byte, error) {
	var message []byte
	var err error
	switch M := NM.(type) {
	case *Header:
		message, err = proto.Marshal(
			&RecordMessage{
				Record: &RecordMessage_Header{
					Header: M,
				}})
	case *Key:
		message, err = proto.Marshal(
			&RecordMessage{
				Record: &RecordMessage_Key{
					Key: M,
				}})
	case *Command:
		message, err = proto.Marshal(
			&RecordMessage{
				Record: &RecordMessage_Command{
					Command: M,
				}})
	case *Terminal:
		message, err = proto.Marshal(
			&RecordMessage{
				Record: &RecordMessage_Terminal{
					Terminal: M,
				}})
	}
	if err != nil {
		return []byte{}, err
	}
	count := make([]byte, 4)
	binary.LittleEndian.PutUint32(count, uint32(len(message)))
	return append(count, message...), nil
}

func ScanMessages(data []byte, atEOF bool) (advance int, token []byte, err error) {
	//Check that we have enough data to read message length. 4 bytes
	if !atEOF && len(data) > 4 {
		messageSize := int(binary.LittleEndian.Uint32(data[0:4]))
		//Check if we have the whole message
		if len(data) >= messageSize+4 {
			return messageSize + 4, data[4 : messageSize+4], nil
		}
	}
	// Request more data.
	return 0, nil, nil
}
