package main

//Buffer data and send in http TCP/IP packets
//packet structure


import (
	"io"
	"fmt"
	"time"
	"net/http"
	"golang.org/x/net/websocket"
)

func auth(w http.ResponseWriter, r *http.Request) {
	// TODO check authentication and assign token
}

func open(ws *websocket.Conn) {
	fmt.Println(ws.Request())
	io.Copy(ws, ws)
	// TODO check token is valid

	// Open assoociated cache file

	// Copy data into cache file
}

func StartServer() error {
	http.HandleFunc("/connect", auth)
	http.Handle("/open", websocket.Handler(open))
	//return http.ListenAndServeTLS(":443", "cert", "key", newConn)
	return http.ListenAndServe(":8080", nil)
}

func main() {
	go StartServer()
	testConnect()
	time.Sleep(time.Minute)
}

func testConnect() {
	origin := "http://localhost/"
	url := "ws://localhost:8080/open"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		panic(err)
	}
	if _, err := ws.Write([]byte("hello, world!\n")); err != nil {
	    panic(err)
	}
	var msg = make([]byte, 512)
	var n int
	if n, err = ws.Read(msg); err != nil {
	    panic(err)
	}
	fmt.Printf("Received: %s.\n", msg[:n])
}