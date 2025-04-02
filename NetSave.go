package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"

	pb "buchanan/recorder/pb"
	//timestamp "github.com/golang/protobuf/ptypes/timestamp"
)

//TODO add session resume
//when connecting return unique videoID
//videoID can be used for playing recording or appending to a recording

//TODO add authentication
//Server should write message to disk and then create database entry linking to file

func printHeader(h *pb.Header) error {
	fmt.Println(*h)
	return nil
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

	readHeader := false

	c := make(chan []byte)
	go save2File("out", c)
	S := bufio.NewScanner(conn)
	S.Split(pb.ScanMessages)

	for S.Scan() {
		//Check if message is a header and print it
		if WM, err := pb.ReadMessage(S.Bytes()); err == nil {
			if WM.MessageType == "header" {
				readHeader = true
				//TODO create unique videoID
			}
			WM.Do(printHeader)
		}
		if readHeader == false {
			fmt.Println("ERROR message received before header")
			return
		}
		c <- S.Bytes()
	}
	if err := S.Err(); err != nil {
		panic(err)
	}
}

func save2File(path string, c <-chan []byte) {
	outfile, err := os.Create(path)
	if err != nil {
		fmt.Println("ERROR could not create savefile")
		return
	}
	for M := range c {
		count := make([]byte, 4)
		binary.LittleEndian.PutUint32(count, uint32(len(M)))
		outfile.Write(append(count, M...))
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
