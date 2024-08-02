package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/bcrypt"
)

const (
	defaultTokenLength = 32
)

func RandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)

	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}

	return b, nil
}

func RandomString(s int) (string, error) {
	b, err := RandomBytes(s)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", sha256.Sum256(b)), err
}

func Nonce() (string, error) {
	value := make([]byte, 12)

	if _, err := io.ReadFull(rand.Reader, value); err != nil {
		return "", fmt.Errorf("failed to generate random nonce: %w", err)
	}

	return base64.StdEncoding.EncodeToString(value), nil
}

type Token struct {
	raw string
}

func NewToken(raw string) *Token {
	return &Token{
		raw: raw,
	}
}

func RandomToken() (*Token, error) {
	raw, err := RandomString(defaultTokenLength)

	if err != nil {
		return nil, fmt.Errorf("failed to generate random token: %w", err)
	}

	return NewToken(raw), nil
}

func (t *Token) String() string {
	return t.raw
}

func (t *Token) Digest() (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(t.raw), bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("failed to secure random token: %w", err)
	}

	return base64.URLEncoding.EncodeToString(hash), nil
}

func (t *Token) Validate(digest string) error {
	hash, err := base64.URLEncoding.DecodeString(digest)

	if err != nil {
		return fmt.Errorf("failed to decode token: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(hash, []byte(t.raw)); err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	return nil
}
