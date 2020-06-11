package main

import (
	"fmt"
	"mobchat/config"
	"mobchat/node"
	"mobchat/node/commands"
	"os"
	"strings"
)

type addr struct {
	address string
	port    string
}

func getCheckin() []addr {
	checkin := strings.Split(config.Attr("checkin"), ",")
	addresses := make([]addr, len(checkin))
	for i := range checkin {
		a := strings.Split(checkin[i], ":")
		addresses[i] = addr{
			address: a[0],
			port:    a[1],
		}
	}
	return addresses
}

func makeClientConnections() {
	addresses := getCheckin()
	for _, address := range addresses {
		go node.Connect(address.address, address.port)
	}
}

func main() {
	address := commands.Address{}
	fmt.Println("address:", address.Serialize())
	err := node.Initialize()
	if err != nil {
		fmt.Println(err)
		return
	}
	sigs := make(chan os.Signal, 1)
	port := config.Attr("port")
	go node.Listen(port)
	makeClientConnections()

	<-sigs
}
