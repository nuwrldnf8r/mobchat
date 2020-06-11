package user

import (
	"mobchat/encryption"
)

type User struct {
	Key      encryption.Key
	Personas []Persona
}
