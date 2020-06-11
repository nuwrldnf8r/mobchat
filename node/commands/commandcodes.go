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

	//CmdGetIP - asks for the address of a given ID
	CmdGetIP = 0x07

	//CmdRelayMessage - asks to relay the message to the Given ID (can nest messages to relay to en end point)
	CmdRelayMessage = 0x08

	//CmdBroadcastMessage - respond to and broadcast message to all peers
	CmdBroadcastMessage = 0x09

	//CmdPeerConnected - sends to all nodes to update routing tables
	CmdPeerConnected = 0x10

	//CmdPeerDisconnected - sends to all nodes to update routing tables
	CmdPeerDisconnected = 0x11

	//CmdGeneric - is a generic message
	CmdGeneric = 0x12

	//CmdRequestPeerConn - requests a peer connection - if affirmative - returns IP, else returns 0x00
	CmdRequestPeerConn = 0x13

	//CmdRequestPeerConnResp -
	CmdRequestPeerConnResp = 0x14
)
