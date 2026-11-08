package keyring

import (
	"crypto/ecdh"
	"crypto/hpke"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
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

func (k *KeyPair) EncodedKey() (string, error) {
	payload, err := json.Marshal(k)
	if err != nil {
		return "", fmt.Errorf("encode secret key: %w", err)
	}

	return base64.StdEncoding.EncodeToString(payload), nil
}

func FromEncodedKey(secret string) (*KeyPair, error) {
	secret = strings.TrimSpace(secret)

	if secret == "" {
		return nil, fmt.Errorf("Secret Key wajib diisi")
	}

	payload, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, fmt.Errorf("Secret Key tidak valid: encoding Base64 gagal")
	}

	var result KeyPair
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, fmt.Errorf("Secret Key tidak valid: format data salah")
	}

	if strings.TrimSpace(result.PublicKey) == "" {
		return nil, fmt.Errorf("Secret Key tidak valid: data tidak lengkap")
	}

	if strings.TrimSpace(result.PrivateKey) == "" {
		return nil, fmt.Errorf("Secret Key tidak valid: data tidak lengkap")
	}

	return FromKeys(result.PublicKey, result.PrivateKey), nil
}
