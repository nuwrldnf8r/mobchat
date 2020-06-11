package routing

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"sort"
	"strings"
	"sync"
)

//Table - global var for Routing Table
var (
	Table = newRouting()
	mutex = sync.RWMutex{}
)

//Routing -
type Routing struct {
	Nodes map[string]*Node
}

func newRouting() Routing {
	return Routing{
		Nodes: make(map[string]*Node),
	}
}

//AddNode -
func (routing *Routing) AddNode(node *Node) {
	mutex.Lock()
	routing.Nodes[node.IDString()] = node
	mutex.Unlock()
}

//Check - returns a checksum for routing table
func (routing *Routing) Check() []byte {

	//sort
	mutex.Lock()
	ar := make([]string, 0)
	for key := range routing.Nodes {
		ar = append(ar, key)
	}
	mutex.Unlock()
	sort.Strings(ar)

	checkStr := strings.Join(ar, "")
	check := []byte(checkStr)
	h := sha256.New()
	h.Write(check)
	return h.Sum(nil)
}

//RemoveNode -
func (routing *Routing) RemoveNode(node *Node) {
	mutex.Lock()
	delete(routing.Nodes, node.IDString())
	for _, n := range routing.Nodes {
		delete(n.Connections, node.IDString())
	}
	mutex.Unlock()
}

//Serialize -
func (routing *Routing) Serialize() []byte {
	var buff bytes.Buffer
	index := make(map[string][]byte)
	routingLen := uint32(len(routing.Nodes))
	ln := make([]byte, 4)
	binary.BigEndian.PutUint32(ln, routingLen)
	buff.Write(ln)
	cnt := 0
	for _, node := range routing.Nodes {
		buff.Write(node.Serialize())
		idx := uint32(cnt*144 + 4)
		bidx := make([]byte, 4)
		binary.BigEndian.PutUint32(bidx, idx)
		index[node.IDString()] = bidx
		cnt++
	}
	for _, node := range routing.Nodes {
		lenConnections := len(node.Connections)
		buff.WriteByte(byte(lenConnections))
		for _, n := range node.Connections {
			buff.Write(index[n.IDString()])
		}
	}
	return buff.Bytes()
}

//FindRoute -
func (routing *Routing) FindRoute(ID []byte) []Node {
	return nil
}

//DeserializeRouting -
func DeserializeRouting(data []byte) (Routing, error) {
	routingLen := binary.BigEndian.Uint32(data[0:4])
	arNodes := make([]Node, routingLen)
	for i := range arNodes {
		idx := 4 + i*144
		nodeData := data[idx : idx+144]
		node, err := DeserializeNode(nodeData)
		if err != nil {
			return Routing{}, err
		}
		arNodes[i] = node
	}
	cnt := uint32(0)
	idx := routingLen*144 + 4
	for cnt < routingLen {
		ln := uint8(data[idx])
		i := uint8(0)
		idx++
		for i < ln {
			nodeIdx := binary.BigEndian.Uint32(data[idx:4])
			n := arNodes[nodeIdx]
			arNodes[cnt].Connections[n.IDString()] = &n
			idx += 4
			i++
		}

		cnt++
	}
	routing := Routing{
		Nodes: make(map[string]*Node),
	}
	for _, node := range arNodes {
		routing.Nodes[node.IDString()] = &node
	}
	return routing, nil
}
