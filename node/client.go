package node

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"mobchat/config"
	"mobchat/encryption"
	"mobchat/node/commands"
	"net"
	"time"
)

func listen(conn *Connection) {
	for {
		decoder := gob.NewDecoder(conn.c)
		var m []byte
		err := decoder.Decode(&m)
		if err != nil {
			if err == io.EOF {
				fmt.Println("EOF")
				go _connections.RemoveAndRetry(*conn)
			} else {
				fmt.Println(err)
				conn.c.Close()
				go _connections.RemoveAndRetry(*conn)
			}
			sendPeerDisconnected(conn.id)
			break
		}
		HandleMessage(DeserializeMessage(m), conn)

	}
}

func doHandshake(conn *Connection) {
	pubKey := encryption.Key{
		Public: _me.Key.Public,
	}
	timer := time.NewTimer(100 * time.Millisecond)
	<-timer.C
	hs := commands.NewHandshake(_me.ID(), pubKey, _me.Address)
	msg := NewMessage(hs.Serialize(), false)
	fmt.Println("sending handshake")
	conn.sendMessage(msg)
}

//Connect -
func Connect(address string, port string) error {
	if address == "127.0.0.1" && port == config.Attr("port") {
		return errors.New("Cannot connect to self")
	}
	fmt.Println("connecting to " + address + ":" + port)
	c, err := net.Dial("tcp", address+":"+port)
	if err != nil {
		fmt.Println(err)
		return err
	}
	conn := Connection{
		c:      c,
		addr:   c.RemoteAddr(),
		server: true,
	}
	_connections.Add(&conn)
	go conn.startHandshakeTimeout()
	go listen(&conn)
	go doHandshake(&conn)
	return nil
}
