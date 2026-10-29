package keyring

import (
	"crypto/ecdh"
	"crypto/hpke"
	"encoding/hex"
)

var (
	kem  = hpke.DHKEM(ecdh.X25519())
	kdf  = hpke.HKDFSHA256()
	aead = hpke.ChaCha20Poly1305()
)

type KeyPair struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

func GenerateKeyPair() (*KeyPair, error) {
	privateKey, err := kem.GenerateKey()
	if err != nil {
		return nil, err
	}

	privateKeyBytes, err := privateKey.Bytes()
	if err != nil {
		return nil, err
	}

	return &KeyPair{
		PublicKey:  hex.EncodeToString(privateKey.PublicKey().Bytes()),
		PrivateKey: hex.EncodeToString(privateKeyBytes),
	}, nil
}

func FromKeys(publicKey, privateKey string) *KeyPair {
	return &KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}
}

func FromPublicKey(publicKey string) *KeyPair {
	return &KeyPair{
		PublicKey: publicKey,
	}
}
