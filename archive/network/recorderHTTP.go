package main

import (
	"net/http"
	"encoding/json"
	"os"
	"bufio"
	"fmt"
	"encoding/binary"
	"time"
	pb "./pb"
	proto "github.com/golang/protobuf/proto"
)

func Header2String(header pb.Header) []byte {
	if b , err := json.Marshal(struct {
		Version   int      `json:"version"`
		Width     uint16   `json:"width"`
		Height    uint16   `json:"height"`
		Timestamp int64    `json:"timestamp"`
		Idle      float64 `json:"idle_time_limit"`
		Command   string   `json:"command"`
		Title     string  `json:"title"`
		Env       map[string]string
		Theme     struct {
			FG      string `json:"fg"`
			BG      string `json:"bg"`
			Palette string `json:"palette"`
		}
	}{
		Version:   2,
		Width:     uint16(header.Width),
		Height:    uint16(header.Height),
		Timestamp: int64(header.Timestamp.Nanos),
		Idle:      header.Idle,
		Command:   header.Command,
		Title:     header.Title,
		//Env:       header.Env,
		//Theme:     header.Theme,
	}); err == nil {
		return append(b, "\n\r"...)
	} else {
		panic(err)
	}
}

func Command2Strings(command pb.Command) []byte {
	var tag string
	if command.Input {
		tag = "i"
	} else {
		tag = "o"
	}
	var commands []byte
	for _, key := range command.Keystrokes {
		if b, err := json.Marshal([]interface{}{
			time.Duration(key.Offset).Seconds(),
			tag,
			string(key.Key),
		}); err == nil {
			//commands = append(commands, b...)
			commands = append(commands, append(b, "\n\r"...)...)
		} else {
			panic(err)
		}
	}
	return commands
}

func ScanMessages(data []byte, atEOF bool) (advance int, token []byte, err error) {
	//Check that we have enough data to read message length. 4 bytes
	if !atEOF && len(data) > 4 {
		messageSize := int(binary.LittleEndian.Uint32(data[0:4]))
		//Check if we have the whole message
		if len(data) >= messageSize + 4 {
			return messageSize+ 4, data[4:messageSize + 4], nil
		}	
	}
	// Request more data.
	return 0, nil, nil
}

func translateFile(filepath string) ([]byte, error) {
	fh, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	S := bufio.NewScanner(fh)
	S.Split(ScanMessages)

	var data []byte

	for S.Scan() {
		var msg pb.RecordMessage
		if err := proto.Unmarshal(S.Bytes(), &msg); err != nil {
			return nil, err
		}
		if header := msg.GetHeader(); header != nil {
			data = append(data, Header2String(*header)...)
		} else if command := msg.GetCommand(); command != nil {
			data = append(data, Command2Strings(*command)...)
		} else {
			fmt.Println("Unknown message")
		}
	}
	if err := S.Err(); err != nil {
		return nil, err
	}

	return data, nil
}

func Authenticate(w http.ResponseWriter, r *http.Request) {
	//Check credentials
	//Store token in database
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./")))
	http.HandleFunc("/login", Authenticate)
	http.HandleFunc("/lookup_record", RequestRecord)
	http.ListenAndServe(":8080", nil)
}

func RequestRecord(w http.ResponseWriter, r *http.Request) {
	//TODO check token
	id := r.FormValue("recordID")
	if id == "" {
		http.Error(w, "record not found", 404)
		return
	}
	filepath := lookupID(id)
	w.Header().Set("Content-Type", "application/json")
	b, err := translateFile(filepath)
	if err != nil {
		http.Error(w, "", 500)
		return
	}
	w.Write(b)
}

func lookupID(id string) string {
	return "/home/nbuchanan/.recorder/"+id
}