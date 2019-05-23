package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"

	pb "../pb"
	proto "github.com/golang/protobuf/proto"
	//timestamp "github.com/golang/protobuf/ptypes/timestamp"
)

//TODO add authentication
//Server should write message to disk and then create database entry linking to file

func main() {
	c := make(chan *pb.RecordMessage)
	go func() {
		for {
			message := <-c
			if header := message.GetHeader(); header != nil {
				fmt.Println("New Message", header)
			} else if key := message.GetKey(); key != nil {
				fmt.Println("New Message", key)
				//go writeKeystrokes([]*pb.Key{key}, time.Now())
			} else if command := message.GetCommand(); command != nil {
				if !command.Input {
					go writeKeystrokes(command.Keystrokes, time.Now())
				}
			} else {
				fmt.Println("Received unknown message")
			}
		}
	}()

	listener, err := net.Listen("tcp", "127.0.0.1:2110")
	if err != nil {
		panic(err)
	}
	for {
		if conn, err := listener.Accept(); err == nil {
			go handleProtoClient(conn, c)
		} else {
			fmt.Println(err.Error())
		}
	}
}

func handleProtoClient(conn net.Conn, c chan *pb.RecordMessage) {
	defer conn.Close()

	S := bufio.NewScanner(conn)
	S.Split(ScanMessages)

	for S.Scan() {
		msg := new(pb.RecordMessage)
		if err := proto.Unmarshal(S.Bytes(), msg); err != nil {
			panic(err)
		}
		c <- msg
	}
	if err := S.Err(); err != nil {
		panic(err)
	}
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

func writeKeystrokes(keylist []*pb.Key, baseTime time.Time) {
	delay := time.NewTimer(time.Minute) //TODO fix this garbage
	if !delay.Stop() {
		panic("AAHH!!")
	}
	for _, key := range keylist {
		if key.GetInput() {
			continue
		}
		nanoseconds := time.Duration(key.Offset) - time.Since(baseTime)
		delay.Reset(nanoseconds)

		select {
		//case <- Command from stdin (spacebar pauses) This should not break the for
		case <-delay.C:
			os.Stdout.Write(key.Key)
		}
	}
}
