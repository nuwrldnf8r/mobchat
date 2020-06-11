package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"time"
)

//Message -
type Message struct {
	Body      []byte
	Timestamp uint64
}

func client() {
	// connect to the server
	c, err := net.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		reader := bufio.NewReader(os.Stdin)
		btxt, _, _ := reader.ReadLine()
		ts := uint64(time.Now().UnixNano())

		msg := Message{
			Body:      btxt,
			Timestamp: ts,
		}
		encoder := gob.NewEncoder(c)
		encoder.Encode(msg)

	}

}

func main() {
	sigs := make(chan os.Signal, 1)
	client()
	<-sigs
}
