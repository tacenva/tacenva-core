package keyring

import (
	"crypto/hpke"
	"encoding/hex"
)

func (k *KeyPair) Seal(data []byte) (string, error) {
	publicKeyBytes, err := hex.DecodeString(k.PublicKey)
	if err != nil {
		return "", err
	}

	publicKey, err := kem.NewPublicKey(publicKeyBytes)
	if err != nil {
		return "", err
	}

	sealed, err := hpke.Seal(
		publicKey,
		kdf,
		aead,
		nil,
		data,
	)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(sealed), nil
}

func (k *KeyPair) Open(data string) ([]byte, error) {
	privateKeyBytes, err := hex.DecodeString(k.PrivateKey)
	if err != nil {
		return nil, err
	}

	sealed, err := hex.DecodeString(data)
	if err != nil {
		return nil, err
	}

	privateKey, err := kem.NewPrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	return hpke.Open(
		privateKey,
		kdf,
		aead,
		nil,
		sealed,
	)
}
