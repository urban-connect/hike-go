package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

func NewAES(key string, nonce string) (*AES, error) {
	aeskey, err := base64.StdEncoding.DecodeString(key)

	if err != nil {
		return nil, fmt.Errorf("failed to decode key: %w", err)
	}

	aesNonce, err := base64.StdEncoding.DecodeString(nonce)

	if err != nil {
		return nil, fmt.Errorf("failed to decode nonce: %w", err)
	}

	return &AES{
		key:   aeskey,
		nonce: []byte(aesNonce),
	}, nil
}

type AES struct {
	key   []byte
	nonce []byte
}

func (s *AES) Decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)

	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	sealer, err := cipher.NewGCM(block)

	if err != nil {
		return nil, fmt.Errorf("failed to create sealer: %w", err)
	}

	raw, err := sealer.Open(nil, s.nonce, data, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to unseal data: %w", err)
	}

	return raw, nil
}

func (s *AES) Encrypt(raw []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)

	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	sealer, err := cipher.NewGCM(block)

	if err != nil {
		return nil, fmt.Errorf("failed to create sealer: %w", err)
	}

	return sealer.Seal(nil, s.nonce, raw, nil), nil
}
