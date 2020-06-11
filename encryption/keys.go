package encryption

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"math/big"
)

const (
	//EncryptedCypherKeyLen -
	EncryptedCypherKeyLen = 128

	//SigLen -
	SigLen = 128
)

//Key -
type Key struct {
	Private *rsa.PrivateKey
	Public  *rsa.PublicKey
}

func sha(b []byte) []byte {
	h := sha256.New()
	h.Write(b)
	hash := h.Sum(nil)
	return hash
}

func serializePriv(priv *rsa.PrivateKey) []byte {
	if priv == nil {
		return nil
	}
	var buf bytes.Buffer
	d := priv.D.Bytes()
	lenD := make([]byte, 2)
	binary.BigEndian.PutUint16(lenD, uint16(len(d)))
	buf.Write(lenD)
	buf.Write(d)
	for _, prime := range priv.Primes {
		buf.Write(prime.Bytes())
	}
	buf.Write(priv.N.Bytes())

	return buf.Bytes()
}

func deserializePriv(priv []byte) rsa.PrivateKey {
	len := binary.BigEndian.Uint16(priv[0:2])
	d := big.Int{}
	d.SetBytes(priv[2 : 2+len])
	prime1 := big.Int{}
	prime2 := big.Int{}
	prime1.SetBytes(priv[2+len : 2+len+len/2])
	prime2.SetBytes(priv[2+len+len/2 : 2+len*2])
	primes := []*big.Int{&prime1, &prime2}
	n := big.Int{}
	n.SetBytes(priv[2+len*2:])
	key := rsa.PrivateKey{
		D:      &d,
		Primes: primes,
	}
	key.N = &n
	key.E = 65537
	key.PublicKey = rsa.PublicKey{
		N: key.N,
		E: key.E,
	}
	return key
}

func serializePub(pub *rsa.PublicKey) []byte {
	var buf bytes.Buffer
	/*
		e := make([]byte, 8)
		binary.BigEndian.PutUint64(e, uint64(pub.E))
	*/
	n := pub.N.Bytes()
	//buf.Write(e)
	buf.Write(n)
	return buf.Bytes()
}

func deserializePub(pub []byte) rsa.PublicKey {
	//e := binary.BigEndian.Uint64(pub[0:8])
	i := big.Int{}
	n := i.SetBytes(pub[:])
	pubKey := rsa.PublicKey{
		N: n,
		E: 65537,
	}
	return pubKey
}

func check(data []byte) (bool, []byte) {
	checksum := sha(data[2:])[0:2]
	if checksum[0] != data[0] || checksum[1] != data[1] {
		return false, nil
	}
	return true, data[2:]
}

//Serialize -
func (key *Key) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	if key.Private == nil {
		if key.Public == nil {
			return nil, errors.New("Need to have either a public or private key. Both can't be nil")
		}
		buf.Write([]byte{0x00, 0x00})
		buf.Write(serializePub(key.Public))
	} else {
		priv := serializePriv(key.Private)
		buf.Write(priv)
	}
	result := buf.Bytes()
	checksum := sha(result)
	result = append(checksum[0:2], result[:]...)
	return result, nil
}

//Deserialize -
func Deserialize(key []byte) (Key, error) {
	valid, checked := check(key)
	if !valid {
		return Key{}, errors.New("Checksum not valid")
	}
	len := binary.BigEndian.Uint16(checked[0:2])
	if len == 0 {
		pubKey := deserializePub(checked[2:])
		return Key{
			Public: &pubKey,
		}, nil
	}
	priv := deserializePriv(checked)

	err := priv.Validate()
	if err != nil {
		return Key{}, err
	}

	priv.Precompute()
	return Key{
		Private: &priv,
		Public:  &priv.PublicKey,
	}, nil
}

//Generate - generates a new Key
func Generate(bits int) (Key, error) {
	privkey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return Key{}, nil
	}
	key := Key{
		Private: privkey,
		Public:  &privkey.PublicKey,
	}
	return key, nil
}

//GetCipherKey -
func GetCipherKey() []byte {
	b := make([]byte, 32)
	rand.Read(b)
	return b
}

//EncryptCypherKey -
func EncryptCypherKey(recipientKey Key, cipherkey []byte) ([]byte, error) {
	hash := sha256.New()
	label := []byte("")
	return rsa.EncryptOAEP(hash, rand.Reader, recipientKey.Public, cipherkey, label)
}

//EncryptMsg -
func EncryptMsg(cipherKey []byte, msg []byte) ([]byte, error) {
	key := sha(cipherKey)
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	encrypted := gcm.Seal(nonce, nonce, msg, nil)
	return encrypted, nil
}

//Encrypt -
func Encrypt(recipientKeys []Key, msg []byte, cipherKey []byte) ([]byte, error) {
	numKeys := make([]byte, 2)
	binary.BigEndian.PutUint16(numKeys, uint16(len(recipientKeys)))
	var buf bytes.Buffer

	buf.Write(numKeys)

	for _, recipientKey := range recipientKeys {
		encryptedCypherKey, err := EncryptCypherKey(recipientKey, cipherKey)
		if err != nil {
			return nil, err
		}
		buf.Write(encryptedCypherKey)
	}

	encryptedMessage, err := EncryptMsg(cipherKey, msg)
	if err != nil {
		return nil, err
	}

	buf.Write(encryptedMessage)

	return buf.Bytes(), nil
}

//Sign -
func Sign(senderKey Key, msg []byte) ([]byte, error) {

	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthEqualsHash
	hash := crypto.SHA256
	pssh := hash.New()
	pssh.Write(msg)
	hashed := pssh.Sum(nil)
	sig, err := rsa.SignPSS(
		rand.Reader,
		senderKey.Private,
		hash,
		hashed,
		&opts,
	)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

//DecryptCypher -
func DecryptCypher(recipientKey Key, encryptedCypherKey []byte) ([]byte, error) {
	hash := sha256.New()
	label := []byte("")
	cypher, err := rsa.DecryptOAEP(
		hash,
		rand.Reader,
		recipientKey.Private,
		encryptedCypherKey,
		label,
	)
	if err != nil {
		return nil, err
	}
	return cypher, nil
}

//DecryptMessage -
func DecryptMessage(key []byte, encryptedMsg []byte) ([]byte, error) {
	c, err := aes.NewCipher(sha(key))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	nonce, encryptedMsg := encryptedMsg[:nonceSize], encryptedMsg[nonceSize:]
	msg, err := gcm.Open(nil, nonce, encryptedMsg, nil)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

//Decrypt -
func Decrypt(key Key, msg []byte) ([]byte, error) {
	cypher, err := DecryptCypher(key, msg[0:EncryptedCypherKeyLen])
	if err != nil {
		return nil, errors.New("Key cannot unlock this message")
	}
	return DecryptMessage(cypher, msg)
}

//ValidateSig -
func ValidateSig(senderKey Key, sig []byte, msg []byte) bool {
	if senderKey.Public == nil || len(sig) < SigLen {
		return false
	}
	hash := crypto.SHA256
	pssh := hash.New()
	pssh.Write(msg)
	hashed := pssh.Sum(nil)
	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthEqualsHash
	err := rsa.VerifyPSS(
		senderKey.Public,
		hash,
		hashed,
		sig,
		&opts,
	)
	return err == nil
}
