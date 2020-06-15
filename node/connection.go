package node

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"mobchat/encryption"
	"mobchat/node/routing"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	msgMax           = 100
	retryMax         = 10
	handshakeTimeout = 10
)

var (
	mutex = sync.RWMutex{}
)

//Connection -
type Connection struct {
	c              net.Conn
	addr           net.Addr
	server         bool
	messageIds     []byte
	handshake      bool
	sentGetRouting bool //prevents malicious getrouting messages
	isPeer         bool
	id             []byte
	pubKey         encryption.Key
	timer          *time.Timer
}

//Connections -
type Connections struct {
	_lst map[string]*Connection
}

func (cons *Connections) countIncoming() int64 {
	cnt := int64(0)
	mutex.Lock()
	for _, c := range _connections._lst {
		if !c.server {
			cnt++
		}
	}
	mutex.Unlock()
	return cnt
}

func (cons *Connections) countOutgoing() int64 {
	cnt := int64(0)
	mutex.Lock()
	for _, c := range _connections._lst {
		if c.server {
			cnt++
		}
	}
	mutex.Unlock()
	return cnt
}

func (con *Connection) sendMessage(msg Message) error {
	encoder := gob.NewEncoder(con.c)
	err := encoder.Encode(msg.Serialize())
	return err
}

func (con *Connection) startHandshakeTimeout() {
	dur, _ := time.ParseDuration(strconv.FormatInt(handshakeTimeout, 10) + "s")
	con.timer = time.NewTimer(dur)

	<-con.timer.C
	if !con.handshake {
		con.c.Close()
	}
}

func (con *Connection) startTimeout() {
	dur, _ := time.ParseDuration(strconv.FormatInt(handshakeTimeout, 30) + "s")
	con.timer = time.NewTimer(dur)

	<-con.timer.C
	if !con.handshake {
		con.c.Close()
	}
}

func (con *Connection) stopHandshakeTimeout() {
	con.handshake = true
	con.timer.Stop()
}

func (con *Connection) stopTimeout() {
	con.timer.Stop()
}

func (con *Connection) close() error {
	return con.c.Close()
}

func (con *Connection) addMessageID(messageID []byte) error {
	mutex.Lock()
	if bytes.Contains(con.messageIds, messageID) {
		mutex.Unlock()
		return errors.New("messageID already exists")
	}
	con.messageIds = append(con.messageIds, messageID...)
	if len(con.messageIds) > 32*msgMax {
		con.messageIds = con.messageIds[32:]
	}
	mutex.Unlock()
	return nil
}

//Add -
func (cons *Connections) Add(con *Connection) {
	mutex.Lock()
	_connections._lst[con.addr.String()] = con
	mutex.Unlock()
}

//Remove -
func (cons *Connections) Remove(con Connection) {
	mutex.Lock()
	delete(_connections._lst, con.addr.String())
	fmt.Println("removed", con.addr.String())
	mutex.Unlock()
}

//Contains - checks to see if a node is a peer
func (cons *Connections) Contains(node *routing.Node, lock bool) bool {
	if lock {
		mutex.Lock()
	}
	for _, con := range cons._lst {
		if con.addr.String() == node.Address.String() {
			if lock {
				mutex.Unlock()
			}
			return true
		}
	}
	if lock {
		mutex.Unlock()
	}

	return false
}

//RemoveAndRetry -
func (cons *Connections) RemoveAndRetry(con Connection) {
	fmt.Println("Removing " + con.addr.String())
	cons.Remove(con)
	if !con.isPeer {
		return
	}
	retries := 0
	for retries < retryMax {
		retries++
		secs, _ := time.ParseDuration(strconv.FormatInt(int64(retries*10), 10) + "s")
		fmt.Println("Retrying in", secs)
		timer := time.NewTimer(secs)
		<-timer.C
		mutex.Lock()
		_, exists := _connections._lst[con.addr.String()]
		if !exists {
			a := strings.Split(con.addr.String(), ":")
			go Connect(a[0], a[1])
			mutex.Unlock()
			break
		}
		mutex.Unlock()
	}
}

//SendMessage -
func (cons *Connections) SendMessage(msg Message) {
	mutex.Lock()
	for _, con := range cons._lst {
		if con.isPeer {
			go con.sendMessage(msg)
		}
	}
	mutex.Unlock()
}
