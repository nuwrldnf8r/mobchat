package node

import (
	"fmt"
	"mobchat/config"
	"mobchat/node/routing"
	"strconv"
)

func findPeers() {
	maxOutgoing, _ := strconv.ParseInt(config.Attr("maxoutgoing"), 10, 64)
	if _connections.countOutgoing() >= maxOutgoing {
		return
	}
	mutex.Lock()
	fmt.Println(len(routing.Table.Nodes))
	for key := range routing.Table.Nodes {
		node := routing.Table.Nodes[key]
		if _connections.Contains(node, false) || node.RequestedPeer || _me.Address.String() == node.Address.String() {
			continue
		}
		if node.IsServer() {

			node.RequestedPeer = true
			go Connect(node.Address.IP, node.Address.Port)

			break

		}
	}
	mutex.Unlock()
}
