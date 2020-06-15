package node

import (
	"crypto/sha256"
	"fmt"
	"mobchat/config"
	"mobchat/encryption"
	"mobchat/node/commands"
	"mobchat/node/routing"
)

//Me -
type Me struct {
	Key          encryption.Key
	PublicServer bool
	Address      commands.Address
}

var (
	_connections    Connections
	_initialized    bool = false
	_me             Me
	_initialRouting = false
)

//Initialize - initializes node
func Initialize() error {
	fmt.Println("initializing")
	_connections = Connections{_lst: make(map[string]*Connection)}
	_initialized = true
	key, err := encryption.Generate(1024)
	if err != nil {
		fmt.Println(err)
		return err
	}
	_me = Me{
		Key:     key,
		Address: commands.NewAddress(config.Attr("address"), config.Attr("port")),
	}
	node := routing.NewNode(encryption.Key{Public: _me.Key.Public}, _me.Address, nil)
	routing.Table.AddNode(&node)
	return nil
}

//ID -
func (me *Me) ID() []byte {
	h := sha256.New()
	id := me.Key.Public.N.Bytes()
	h.Write(id)
	return h.Sum(nil)
}
