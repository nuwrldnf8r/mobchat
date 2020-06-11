package commands

import (
	"bytes"
	"encoding/binary"
	"strconv"
	"strings"
)

//Address -
type Address struct {
	IP   string
	Port string
}

//String -
func (address *Address) String() string {
	return address.IP + ":" + address.Port
}

func serializeIP(ip string) []byte {
	ar := strings.Split(ip, ".")
	var buff bytes.Buffer
	for _, s := range ar {
		i, _ := strconv.ParseUint(s, 10, 8)
		buff.WriteByte(uint8(i))
	}
	return buff.Bytes()
}

func deserializeIP(ip []byte) string {
	ar := make([]string, 4)
	for i, b := range ip {
		ar[i] = strconv.FormatUint(uint64(uint8(b)), 10)
	}
	return strings.Join(ar, ".")
}

//Serialize -
func (address *Address) Serialize() []byte {
	if len(address.IP) == 0 && len(address.Port) == 0 {
		return make([]byte, 12)
	}
	var buff bytes.Buffer
	port, _ := strconv.ParseUint(address.Port, 10, 64)
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, port)
	buff.Write(serializeIP(address.IP))
	buff.Write(b)
	return buff.Bytes()
}

//DeserializeAddress -
func DeserializeAddress(b []byte) (Address, error) {
	ar := make([]string, 4)
	for i := range ar {
		ar[i] = strconv.FormatUint(uint64(uint8(b[i])), 10)
	}
	return Address{
		IP:   strings.Join(ar, "."),
		Port: strconv.FormatUint(binary.BigEndian.Uint64(b[4:]), 10),
	}, nil
}

//NewAddress -
func NewAddress(IP, port string) Address {

	return Address{
		IP:   IP,
		Port: port,
	}
}
