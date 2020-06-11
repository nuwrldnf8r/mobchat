package main

import (
	"fmt"
	"mobchat/node/commands"
)

func main() {
	address := commands.Address{}
	b := address.Serialize()
	address2, _ := commands.DeserializeAddress(b)
	fmt.Println(address2.IP)
}
