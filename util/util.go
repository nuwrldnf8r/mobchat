package util

import "encoding/hex"

//ToHexString - converts from a []byte to a hex string
func ToHexString(data []byte) string {
	dst := make([]byte, hex.EncodedLen(32))
	hex.Encode(dst, data)
	return string(dst)
}
