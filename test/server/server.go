package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
)

func server() {
	// listen on a port

	ln, err := net.Listen("tcp", ":9999")

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Listening on 9999")
	for {
		// accept a connection
		c, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		// handle the connection
		go handleServerConnection(c)
	}
}

func handleServerConnection(conn net.Conn) {
	// receive the message
	fmt.Println(conn.RemoteAddr().String(), "connected")
	//reader := bufio.NewReader(conn)
	for {

		decoder := gob.NewDecoder(conn)
		var m []byte
		err := decoder.Decode(&m)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Println(string(m))

	}
}

func main() {
	sigs := make(chan os.Signal, 1)
	go server()

	<-sigs
	fmt.Println("goodbye")
}
