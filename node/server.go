package node

import (
	"encoding/gob"
	"fmt"
	"io"
	"net"
)

//Listen - starts listening to the given port for incoming connections
func Listen(port string) error {
	if !_initialized {
		Initialize()
	}
	// listen on a port

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Listening on", port)
	for {
		// accept a connection
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		// handle the connection
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {

	// receive the message
	fmt.Println(conn.RemoteAddr().String(), "connected")
	c := Connection{
		c:      conn,
		addr:   conn.RemoteAddr(),
		server: false,
	}
	_connections.Add(&c)
	go c.startHandshakeTimeout()
	//c.sendMessage([]byte("hello"))
	for {
		decoder := gob.NewDecoder(conn)
		var m []byte
		err := decoder.Decode(&m)
		if err != nil {
			if err == io.EOF {
				_connections.Remove(c)
			} else {
				fmt.Println(err)
				conn.Close()
				_connections.Remove(c)
			}
			break
		}
		go HandleMessage(DeserializeMessage(m), &c)

	}

}
