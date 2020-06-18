package routing

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"mobchat/node/commands"
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
	_, exists := routing.Nodes[node.IDString()]
	if !exists {
		routing.Nodes[node.IDString()] = node
	}
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

//Compare - compares a returned check with this routing
func (routing *Routing) Compare(check []byte) bool {
	return bytes.Compare(routing.Check(), check) == 0
}

//FindNodeByAddress -
func (routing *Routing) FindNodeByAddress(addr commands.Address) *Node {
	_addr := addr.String()
	for _, node := range routing.Nodes {
		if node.Address.String() == _addr {
			return node
		}
	}
	return nil
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
	//write routing length
	routingLen := uint32(len(routing.Nodes))
	ln := make([]byte, 4)
	binary.BigEndian.PutUint32(ln, routingLen)
	buff.Write(ln)

	//write nodes
	cnt := 0
	for _, node := range routing.Nodes {
		buff.Write(node.Serialize())
		//add index of this node to the index map
		//to map out connections later
		idx := uint32(cnt)
		bidx := make([]byte, 4)
		binary.BigEndian.PutUint32(bidx, idx)
		index[node.IDString()] = bidx
		cnt++
	}
	//write connections for each node by index
	for _, node := range routing.Nodes {
		lenConnections := len(node.Connections)
		buff.WriteByte(byte(lenConnections))
		for _, n := range node.Connections {
			buff.Write(index[n.IDString()])
		}
	}
	return buff.Bytes()
}

//DeserializeRouting -
func DeserializeRouting(data []byte) (Routing, error) {
	//get the routing length (number of nodes)
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
	//add indexing
	cnt := uint32(0)
	idx := routingLen*144 + 4
	for cnt < routingLen {
		ln := uint8(data[idx])
		i := uint8(0)
		arNodes[cnt].Connections = make(map[string]*Node)
		idx++
		for i < ln {
			nodeIdx := binary.BigEndian.Uint32(data[idx : idx+4])
			n := arNodes[nodeIdx]
			arNodes[cnt].AddConnection(&n)
			idx += 4
			i++
		}

		cnt++
	}
	routing := Routing{
		Nodes: make(map[string]*Node),
	}
	for i := range arNodes {
		n := arNodes[i]
		routing.Nodes[n.IDString()] = &n
	}
	return routing, nil
}
