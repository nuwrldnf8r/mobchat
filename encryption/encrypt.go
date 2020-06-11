package encryption

import (
	"bytes"
	"errors"
)

//SignAndEncrypt - signs and encrypts the message body.
//Returns the message body together with the signature
func SignAndEncrypt(body []byte, senderKey Key, recipientKey Key) ([]byte, error) {
	cipher := GetCipherKey()
	encryptedCypher, err := EncryptCypherKey(recipientKey, cipher)
	if err != nil {
		return nil, err
	}
	encryptedBody, err := EncryptMsg(cipher, body)
	if err != nil {
		return nil, err
	}
	var buff bytes.Buffer
	buff.Write(encryptedCypher)
	buff.Write(encryptedBody)
	msg := buff.Bytes()
	sig, err := Sign(senderKey, msg)
	if err != nil {
		return nil, err
	}
	msg = append(sig, msg...)
	return msg, nil
}

//ValidateSigAndDecrypt -
func ValidateSigAndDecrypt(senderKey Key, recipientKey Key, body []byte) ([]byte, error) {
	sig := body[0:SigLen]
	msg := body[SigLen:]
	isValid := ValidateSig(senderKey, sig, msg)
	if !isValid {
		return nil, errors.New("Invalid signature")
	}
	return Decrypt(recipientKey, msg)
}
