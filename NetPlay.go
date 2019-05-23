package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"

	pb "./pb"
	//timestamp "github.com/golang/protobuf/ptypes/timestamp"
)

//TODO add session resume
//TODO add authentication
//Server should write message to disk and then create database entry linking to file

func printHeader(h *pb.Header) error {
	fmt.Println(*h)
	return nil
}

func handleMessage(message []byte) {
	WM, err := readMessage(message)
	if err != nil {
		fmt.Println("ERROR", err.Error())
		return
	}
	if err := WM.Do(printHeader); err != nil {
		fmt.Println("ERROR", err.Error())
	}
	WM.Do(func(m *pb.Key) error {
		fmt.Println("New Message", m)
		return nil
	})
	switch WM.MessageType {
	case "command":
		fmt.Println("New Message", WM.Command)
	case "terminal":
		fmt.Println("New Message", WM.Terminal)
	}
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:2110")
	if err != nil {
		panic(err)
	}
	for {
		if conn, err := listener.Accept(); err == nil {
			go handleProtoClient(conn)
		} else {
			fmt.Println(err.Error())
		}
	}
}

func handleProtoClient(conn net.Conn) {
	defer conn.Close()
	//Recover if bufio buffer overrun
	defer func() { recover() }()

	S := bufio.NewScanner(conn)
	S.Split(ScanMessages)

	for S.Scan() {
		handleMessage(S.Bytes())
	}
	if err := S.Err(); err != nil {
		panic(err)
	}
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
