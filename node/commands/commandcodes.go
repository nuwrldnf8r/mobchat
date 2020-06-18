package commands

const (
	//Version -
	Version = 0x01

	/*****commands******/

	//CmdHandshake -
	CmdHandshake = 0x01

	//CmdHandshakeResp -
	CmdHandshakeResp = 0x02

	//CmdCheckRouting - requests a hash of the routing table to compare to own
	CmdCheckRouting = 0x03

	//CmdCheckRoutingResp - response to the check routing request
	CmdCheckRoutingResp = 0x04

	//CmdGetRouting - requests routing table
	CmdGetRouting = 0x05

	//CmdGetRoutingResp - response to resquest routing table
	CmdGetRoutingResp = 0x06

	//CmdGetRoute - asks peer for the route to a certain ID
	CmdGetRoute = 0x07

	//CmdGetRouteResp - asks peer for the route to a certain ID
	CmdGetRouteResp = 0x08

	//CmdRelayMessage - asks to relay the message to the Given ID (can nest messages to relay to en end point)
	CmdRelayMessage = 0x09

	//CmdBroadcastMessage - respond to and broadcast message to all peers
	CmdBroadcastMessage = 0x10

	//CmdPeerConnected - sends to all nodes to update routing tables
	CmdPeerConnected = 0x11

	//CmdPeerDisconnected - sends to all nodes to update routing tables
	CmdPeerDisconnected = 0x12

	//CmdGeneric - is a generic message
	CmdGeneric = 0x13
)
