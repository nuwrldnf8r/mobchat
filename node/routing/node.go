package routing

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"mobchat/encryption"
	"mobchat/node/commands"
)

//Node -
type Node struct {
	PubKey      encryption.Key
	Address     commands.Address
	Connections map[string]*Node
}

//ID -
func (node *Node) ID() []byte {
	h := sha256.New()
	id := node.PubKey.Public.N.Bytes()
	h.Write(id)
	return h.Sum(nil)
}

//IsServer - shows whether this noe can accept connections
func (node *Node) IsServer() bool {
	return node.Address.IP == "0.0.0.0"
}

//Serialize -
func (node *Node) Serialize() []byte {
	var buff bytes.Buffer
	//buff.Write(node.ID)
	pubKey, _ := node.PubKey.Serialize()
	buff.Write(pubKey)
	buff.Write(node.Address.Serialize()) //len 12

	return buff.Bytes()
	//total len 32 + 12
}

//AddConnection -
func (node *Node) AddConnection(n *Node) {
	mutex.Lock()
	node.Connections[n.IDString()] = n
	mutex.Unlock()
}

//RemoveConnection -
func (node *Node) RemoveConnection(n *Node) {
	mutex.Lock()
	delete(node.Connections, n.IDString())
	mutex.Unlock()
}

//NewNode -
func NewNode(pubKey encryption.Key, address commands.Address, connections []*Node) Node {
	node := Node{
		Address: address,
		PubKey: encryption.Key{
			Public: pubKey.Public,
		},
		Connections: make(map[string]*Node),
	}
	for _, n := range connections {
		node.AddConnection(n)
	}
	return node
}

//IDString -
func (node *Node) IDString() string {
	dst := make([]byte, hex.EncodedLen(32))
	hex.Encode(dst, node.ID())
	return string(dst)
}

//DeserializeNode -
func DeserializeNode(data []byte) (Node, error) {
	//132
	if len(data) != 144 {
		return Node{}, errors.New("Invalid data - length needs to be 176 bytes")
	}
	pubKey, err := encryption.Deserialize(data[0:132])
	if err != nil {
		return Node{}, err
	}
	address, err := commands.DeserializeAddress(data[132:])
	if err != nil {
		return Node{}, err
	}
	return Node{
		PubKey:  pubKey,
		Address: address,
	}, nil
}
