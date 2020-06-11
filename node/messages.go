package node

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"mobchat/config"
	"mobchat/encryption"
	"mobchat/node/commands"
	"mobchat/node/routing"
	"strconv"
	"time"
)

//Message -
type Message struct {
	Body      []byte
	Timestamp uint64
	Encrypted bool
}

//NewMessage -
func NewMessage(body []byte, encrypted bool) Message {
	return Message{
		Body:      body,
		Encrypted: encrypted,
		Timestamp: uint64(time.Now().UnixNano()),
	}
}

//ID -
func (m *Message) ID() []byte {
	if m.Timestamp == 0 {
		m.Timestamp = uint64(time.Now().UnixNano())
	}
	var buff bytes.Buffer
	buff.Write(m.Body)
	timestamp := make([]byte, 8)
	binary.BigEndian.PutUint64(timestamp, m.Timestamp)
	buff.Write(timestamp)
	h := sha256.New()
	h.Write(buff.Bytes())
	return h.Sum(nil)
}

//HandleMessage -
func HandleMessage(msg Message, con *Connection) {
	err := con.addMessageID(msg.ID())
	if err != nil {
		return
	}
	//for now we ignore version
	var body []byte
	if msg.Encrypted {
		body, err = encryption.ValidateSigAndDecrypt(con.pubKey, _me.Key, msg.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		body = msg.Body
	}
	cmd := body[1]
	switch cmd {
	case commands.CmdHandshake:
		hs, err := commands.DeserializeHandshake(body)
		if err != nil {
			fmt.Println(err)
			//respond with error message
		}
		handleHandshake(hs, con)
		break
	case commands.CmdHandshakeResp:
		fmt.Println("body:", body)
		hsr, err := commands.DeserializeHandshakeResponse(body)
		fmt.Println("hsr:", hsr)
		if err != nil {
			fmt.Println(err)
			//respond with error message
		}
		handleHandshakeResp(hsr, con)
		break
	case commands.CmdCheckRouting:
		handleRoutingCheck(con)
		break
	case commands.CmdCheckRoutingResp:
		handleRoutingCheckResp(body[2:], con)
		break
	case commands.CmdGetRouting:
		handleGetRouting(con)
		break
	case commands.CmdGetRoutingResp:
		handleGetRoutingResp(body[2:], con)
		break
	default:
		fmt.Println("Junk message")
		//is junk message - need to decide how to respond
	}
}

func handleHandshake(hs commands.Handshake, con *Connection) {
	mutex.Lock()
	isConnection := true
	con.id = hs.ID
	con.pubKey = hs.PubKey
	mutex.Unlock()
	//check if any connections available
	maxIncoming, _ := strconv.ParseInt(config.Attr("maxincoming"), 10, 64)
	address := commands.NewAddress(config.Attr("address"), config.Attr("port"))
	if _connections.countIncoming() >= maxIncoming {
		address = commands.Address{}
		isConnection = false
	} else {
		con.isPeer = true
	}
	pubKey := encryption.Key{Public: _me.Key.Public}
	s, _ := pubKey.Serialize()
	fmt.Println("serialized key", s)
	fmt.Println("id", _me.ID())
	hsr := commands.NewHandshakeResponse(_me.ID(), pubKey, address)

	msg := NewMessage(hsr.Serialize(), false)
	mutex.Lock()
	con.stopHandshakeTimeout()
	err := con.sendMessage(msg)
	if err != nil {
		fmt.Println(err)
	}
	mutex.Unlock()
	if isConnection {
		n := routing.NewNode(hs.PubKey, hs.Address, nil)
		routing.Table.AddNode(&n)
	}
}

func handleHandshakeResp(hsr commands.HandshakeResponse, con *Connection) {
	con.stopHandshakeTimeout()
	if hsr.IsConnection() {
		n := routing.NewNode(hsr.PubKey, hsr.Address, nil)
		routing.Table.AddNode(&n)
		con.isPeer = true
	}
	//do routing check
	routingCheck := []byte{commands.Version, commands.CmdCheckRouting}
	msg := NewMessage(routingCheck, false)
	err := con.sendMessage(msg)
	if err != nil {
		fmt.Println(err)
	}
}

func handleRoutingCheck(con *Connection) {
	check := routing.Table.Check()
	cmd := []byte{commands.Version, commands.CmdCheckRoutingResp}
	body := append(cmd, check...)
	msg := NewMessage(body, false)
	err := con.sendMessage(msg)
	if err != nil {
		fmt.Println(err)
	}
	if !con.isPeer {
		con.startTimeout()
	}
}

func handleRoutingCheckResp(data []byte, con *Connection) {

	//compare with own routing
	if routing.Table.Compare(data) {
		if !con.isPeer {
			con.close()
		}
		return
	}
	cmd := []byte{commands.Version, commands.CmdGetRouting}
	msg := NewMessage(cmd, false)
	err := con.sendMessage(msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	con.sentGetRouting = true
}

func handleGetRouting(con *Connection) {
	con.stopTimeout()
	var buff bytes.Buffer
	buff.Write([]byte{commands.Version, commands.CmdGetRoutingResp})
	buff.Write(routing.Table.Serialize())
	msg := NewMessage(buff.Bytes(), false)
	err := con.sendMessage(msg)
	if err != nil {
		fmt.Println(err)
	}
	if !con.isPeer {
		con.startTimeout()
	}
}

func handleGetRoutingResp(data []byte, con *Connection) {
	if !con.sentGetRouting {
		con.close()
		return
	}
	route, err := routing.DeserializeRouting(data)
	if err != nil {
		fmt.Println(err)
	}
	for _, n := range route.Nodes {
		routing.Table.AddNode(n)
	}

	go findPeers()

	if !con.isPeer {
		con.close()
	}
}
