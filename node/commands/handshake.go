package commands

import (
	"bytes"
	"fmt"
	"mobchat/encryption"
)

//HandshakeResponse -
type HandshakeResponse struct {
	ID      []byte
	PubKey  encryption.Key
	Address Address
}

//Handshake -
type Handshake struct {
	ID      []byte
	PubKey  encryption.Key
	Address Address
}

//Serialize -
func (hs *Handshake) Serialize() []byte {
	var buff bytes.Buffer
	buff.WriteByte(Version)
	buff.WriteByte(CmdHandshake)
	buff.Write(hs.ID)
	pubKey, _ := hs.PubKey.Serialize()
	buff.Write(pubKey)
	buff.Write(hs.Address.Serialize())
	return buff.Bytes()
}

//NewHandshake -
func NewHandshake(ID []byte, key encryption.Key, address Address) Handshake {
	return Handshake{
		ID: ID,
		PubKey: encryption.Key{
			Public: key.Public,
		},
		Address: address,
	}
}

//DeserializeHandshake -
func DeserializeHandshake(hs []byte) (Handshake, error) {
	id := hs[2:34]
	key, err := encryption.Deserialize(hs[34:166])
	if err != nil {
		return Handshake{}, err
	}
	address, err := DeserializeAddress(hs[166:])
	if err != nil {
		fmt.Println("address error")
		return Handshake{}, err
	}
	return Handshake{
		ID:      id,
		PubKey:  key,
		Address: address,
	}, nil
}

//Serialize -
func (hs *HandshakeResponse) Serialize() []byte {
	var buff bytes.Buffer
	buff.WriteByte(Version)
	buff.WriteByte(CmdHandshakeResp)
	buff.Write(hs.ID)
	pubKey, _ := hs.PubKey.Serialize()
	buff.Write(pubKey)
	buff.Write(hs.Address.Serialize())
	return buff.Bytes()
}

//IsConnection -
func (hs *HandshakeResponse) IsConnection() bool {
	return hs.Address.IP != "0.0.0.0"
}

//NewHandshakeResponse -
func NewHandshakeResponse(ID []byte, pubKey encryption.Key, address Address) HandshakeResponse {
	return HandshakeResponse{
		ID:      ID,
		PubKey:  pubKey,
		Address: address,
	}
}

//DeserializeHandshakeResponse -
func DeserializeHandshakeResponse(hsr []byte) (HandshakeResponse, error) {
	id := hsr[2:34]

	key, err := encryption.Deserialize(hsr[34:166])
	if err != nil {
		return HandshakeResponse{}, err
	}
	address, err := DeserializeAddress(hsr[166:])
	if err != nil {
		fmt.Println("address error")
		return HandshakeResponse{}, err
	}

	return HandshakeResponse{
		ID:      id,
		PubKey:  key,
		Address: address,
	}, nil

}
