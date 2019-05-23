package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	"time"

	pb "../pb"
	proto "github.com/golang/protobuf/proto"
	//timestamp "github.com/golang/protobuf/ptypes/timestamp"
)

//TODO either sort messages/keystrokes or don't support messages out of order and crash

var wg sync.WaitGroup
var baseTime time.Time = time.Now()

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

func play(filepath string) error {
	fh, err := os.Open(filepath)
	if err != nil {
		return err
	}
	S := bufio.NewScanner(fh)
	S.Split(ScanMessages)

	for S.Scan() {
		var msg pb.RecordMessage
		if err := proto.Unmarshal(S.Bytes(), &msg); err != nil {
			fmt.Println(err.Error())
			//return err
		}
		if header := msg.GetHeader(); header != nil {
			fmt.Println(header.GetVersion())
		} else if command := msg.GetCommand(); command != nil {
			if !command.Input {
				wg.Add(1)
				go writeKeystrokes(command.Keystrokes)
			}
		} else if key := msg.GetKey(); key != nil {
			if !key.Input {
				wg.Add(1)
				go writeKeystrokes([]*pb.Key{key})
			}
		} else if term := msg.GetTerminal(); term != nil {
			fmt.Println("Found term")
		} else {
			fmt.Println("Unknown message")
		}
	}
	return S.Err()
}

func writeKeystrokes(keylist []*pb.Key) {
	delay := time.NewTimer(time.Minute) //TODO fix this garbage
	if !delay.Stop() {
		panic("AAHH!!")
	}
	for _, key := range keylist {
		nanoseconds := time.Duration(key.Offset) - time.Since(baseTime)
		delay.Reset(nanoseconds)

		select {
		//case <- Command from stdin (spacebar pauses) This should not break the for
		case <-delay.C:
			os.Stdout.Write(key.Key)
		}
	}
	wg.Done()
}

func main() {
	if err := play(os.Args[1]); err != nil {
		panic(err)
	}
	wg.Wait()
}
