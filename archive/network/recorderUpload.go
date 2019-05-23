package main

import (
	pb "./pb"
	"fmt"
	"os"
	"path/filepath"
	"net"
	"io"
	"time"
	"strconv"
	"encoding/binary"
	proto "github.com/golang/protobuf/proto"
	//timestamp "github.com/golang/protobuf/ptypes/timestamp"
)

// This is the root tree of records recieved
// This folder will have a subfolder for each host that has connected
// Each host subfolder will have records named with the user and time
var recordPath string = "/srv/records/"

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
	fmt.Printf("New record received")
	defer conn.Close()

	//Read header or die
	headerSizeData := make([]byte, 4)
	if n, err := io.ReadFull(conn, headerSizeData); err != nil || n != 4 {
		fmt.Printf("\nCould not read record header. Quitting!\n")
		fmt.Println("DEBUG", err.Error())
		return
	}
	headerSize := int(binary.LittleEndian.Uint32(headerSizeData))
	headerData := make([]byte, headerSize)
	if n, err := io.ReadFull(conn, headerData); err != nil || n != headerSize {
		fmt.Printf("\nCould not read record header. Quitting!\n")
		fmt.Println("DEBUG", err.Error())
		return
	}

	var msg pb.RecordMessage
	if err := proto.Unmarshal(headerData, &msg); err != nil {
		fmt.Printf("\nCound not unmarshal header. Quitting!\n")
		fmt.Println("DEBUG", err.Error())
		return
	}
	header := msg.GetHeader()
	if header == nil {
		fmt.Printf("\nCould not unmarshal header. Quitting!\n")
		return
	}

	fmt.Printf(" from %s. User: %s Date: %s\n", header.Host, header.Username, time.Unix(header.Timestamp.Seconds/1000000000, header.Timestamp.Seconds%1000000000))
	savepath := recordPath + header.Host + "/" + strconv.FormatInt(header.Timestamp.Seconds, 10) + "_" + header.Username + ".rec"
	fmt.Printf("Saving record to: %s\n", savepath)
	if err := os.MkdirAll(filepath.Dir(savepath), 0777); err != nil {
		fmt.Println("ERROR: Cannot create save directory. Quitting!", err.Error())
		return
	}
	fh, err := os.OpenFile(savepath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0444)
	if err != nil {
		if os.IsExist(err) {
			fmt.Println("ERROR: file already exists. Quitting!")
			return
		}
		if os.IsPermission(err) {
			fmt.Println("ERROR: Cannot open file, permission denied.")
			return
		}
		fmt.Println("ERROR: Cannot save record, error opening file.", err.Error())
		return
	}

	defer fh.Close()
	// go Create database entry with header info and filename

	if n, err := fh.Write(headerSizeData); err != nil || n != 4 {
		fmt.Println("ERROR: Could not write to savefile.")
		fmt.Println("DEBUG", err.Error())
		return
	}
	if n, err := fh.Write(headerData); err != nil || n != headerSize {
		fmt.Println("ERROR: Could not write to savefile")
		fmt.Println("DEBUG", err.Error())
		return
	}
	if _, err := io.Copy(fh, conn); err != nil {
		fmt.Println("ERROR: Could not write to savefile")
		fmt.Println("DEBUG", err.Error())
		return
	}
}