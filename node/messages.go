package node

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
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

//Serialize -
func (msg *Message) Serialize() []byte {
	var buff bytes.Buffer
	if msg.Encrypted {
		buff.WriteByte(0x01)
	} else {
		buff.WriteByte(0x00)
	}
	ts := make([]byte, 8)
	binary.BigEndian.PutUint64(ts, msg.Timestamp)
	buff.Write(ts)
	buff.Write(msg.Body)
	return buff.Bytes()
}

//DeserializeMessage -
func DeserializeMessage(data []byte) Message {
	m := Message{}
	if data[0] == 0x01 {
		m.Encrypted = true
	}
	m.Timestamp = binary.BigEndian.Uint64(data[1:9])
	m.Body = data[9:]
	return m
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
		hsr, err := commands.DeserializeHandshakeResponse(body)
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
	case commands.CmdPeerConnected:
		handlePeerConnected(msg)
		break
	case commands.CmdPeerDisconnected:
		handlePeerDisconnected(msg)
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
		con.id = hs.ID

	}
	pubKey := encryption.Key{Public: _me.Key.Public}
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
		go sendConnectionMessage(&n)
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
	mutex.Lock()
	if _initialRouting {
		mutex.Unlock()
		if !con.isPeer {
			fmt.Println("messages.go 151")
			con.close()
		}
		go findPeers()
		return
	}
	mutex.Unlock()
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
			fmt.Println("messages.go 185")
			con.close()
		}
		go findPeers()
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
		fmt.Println("messages.go 281")
		con.close()
		go findPeers()
		return
	}
	fmt.Println("***************************")
	fmt.Println(data)
	fmt.Println("***************************")
	route, err := routing.DeserializeRouting(data)
	if err != nil {
		fmt.Println(err)
	}
	for key := range route.Nodes {
		routing.Table.AddNode(route.Nodes[key])
	}

	if !con.isPeer {
		fmt.Println("messages.go 235")
		con.close()
	}

	go findPeers()
}

func sendConnectionMessage(node *routing.Node) {
	var buff bytes.Buffer
	buff.Write([]byte{commands.Version, commands.CmdPeerConnected})
	buff.Write(_me.ID())
	buff.Write(node.Serialize()) //len 144
	sig, err := encryption.Sign(_me.Key, buff.Bytes())

	if err != nil {
		fmt.Println(err)
		return
	}
	buff.Write(sig)
	msg := NewMessage(buff.Bytes(), false)
	_connections.SendMessage(msg)
}

func handlePeerConnected(msg Message) {
	data := msg.Body[2:]
	dst := make([]byte, hex.EncodedLen(32))
	hex.Encode(dst, data[0:32])
	id := string(dst)
	mutex.Lock()
	node, exists := routing.Table.Nodes[id]
	mutex.Unlock()
	if !exists {
		fmt.Println("Node does not exist")
		//TODO: ask for routing from a peer..
		return
	}
	n, err := routing.DeserializeNode(data[32:176])
	if err != nil {
		fmt.Println(err)
		return
	}
	sig := data[176:]

	if !encryption.ValidateSig(node.PubKey, sig, msg.Body[0:178]) {
		fmt.Println("Invalid sig for peer connection")
		return
	}
	go _connections.SendMessage(msg)

	routing.Table.AddNode(&n)
	node.AddConnection(&n)

}

func handlePeerDisconnected(msg Message) {

	data := msg.Body[2:]
	dst := make([]byte, hex.EncodedLen(32))
	hex.Encode(dst, data[0:32])
	id1 := string(dst)
	mutex.Lock()
	node1, exists := routing.Table.Nodes[id1]
	mutex.Unlock()
	if !exists {
		fmt.Println("Node1 does not exist")
		return
	}
	hex.Encode(dst, data[32:64])
	id2 := string(dst)
	mutex.Lock()
	node2, exists := routing.Table.Nodes[id2]
	mutex.Unlock()
	if !exists {
		fmt.Println("Node2 does not exist")
		return
	}
	sig := data[64:]

	if !encryption.ValidateSig(node1.PubKey, sig, msg.Body[0:66]) {
		fmt.Println("Invalid sig for peer connection")
		return
	}
	go _connections.SendMessage(msg)

	node1.RemoveConnection(node2)

}

func sendPeerDisconnected(id []byte) {
	body := []byte{commands.Version, commands.CmdPeerDisconnected}
	body = append(body, _me.ID()...)
	body = append(body, id...)
	sig, _ := encryption.Sign(_me.Key, body)
	body = append(body, sig...)
	msg := NewMessage(body, false)
	_connections.SendMessage(msg)
}
