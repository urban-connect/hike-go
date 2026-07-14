package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAES(t *testing.T) *AES {
	t.Helper()

	key := make([]byte, 32) // AES-256
	_, err := rand.Read(key)
	require.NoError(t, err)

	nonce := make([]byte, 12) // GCM standard nonce size
	_, err = rand.Read(nonce)
	require.NoError(t, err)

	aes, err := NewAES(
		base64.StdEncoding.EncodeToString(key),
		base64.StdEncoding.EncodeToString(nonce),
	)
	require.NoError(t, err)

	return aes
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	aes := newTestAES(t)

	t.Run("encrypts and decrypts back to original", func(t *testing.T) {
		original := []byte(`{"env":"test","crypto":{"key":"secret"}}`)

		encrypted, err := aes.Encrypt(original)
		require.NoError(t, err)
		assert.NotEqual(t, original, encrypted)

		decrypted, err := aes.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, original, decrypted)
	})

	t.Run("encrypts empty payload", func(t *testing.T) {
		encrypted, err := aes.Encrypt([]byte{})
		require.NoError(t, err)

		decrypted, err := aes.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Empty(t, decrypted)
	})
}

// The fixed nonce is what makes encrypted config files reproducible, so guard it.
func TestEncryptIsDeterministic(t *testing.T) {
	aes := newTestAES(t)

	first, err := aes.Encrypt([]byte("settings"))
	require.NoError(t, err)

	second, err := aes.Encrypt([]byte("settings"))
	require.NoError(t, err)

	assert.Equal(t, first, second)
}

func TestDecryptRejectsTamperedData(t *testing.T) {
	aes := newTestAES(t)

	encrypted, err := aes.Encrypt([]byte("secret"))
	require.NoError(t, err)

	encrypted[0] ^= 0xFF

	_, err = aes.Decrypt(encrypted)
	assert.Error(t, err)
}

func TestNewAESValidation(t *testing.T) {
	nonce := base64.StdEncoding.EncodeToString(make([]byte, 12))

	t.Run("rejects invalid base64 key", func(t *testing.T) {
		_, err := NewAES("not-valid-base64!!!", nonce)
		assert.Error(t, err)
	})

	t.Run("rejects invalid base64 nonce", func(t *testing.T) {
		key := base64.StdEncoding.EncodeToString(make([]byte, 32))
		_, err := NewAES(key, "not-valid-base64!!!")
		assert.Error(t, err)
	})

	t.Run("rejects invalid AES key size at encryption time", func(t *testing.T) {
		aes, err := NewAES(
			base64.StdEncoding.EncodeToString(make([]byte, 10)), // not 16/24/32
			nonce,
		)
		require.NoError(t, err) // NewAES does not validate key size

		_, err = aes.Encrypt([]byte("test"))
		assert.Error(t, err)
	})
}
