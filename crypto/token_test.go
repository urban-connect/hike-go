package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenDigestAndValidate(t *testing.T) {
	token := NewToken("s3cr3t-access-token")

	digest, err := token.Digest()
	require.NoError(t, err)
	assert.NotEmpty(t, digest)

	t.Run("validates the matching token against its digest", func(t *testing.T) {
		require.NoError(t, token.Validate(digest))
	})

	t.Run("rejects a different token", func(t *testing.T) {
		assert.Error(t, NewToken("wrong-token").Validate(digest))
	})

	t.Run("rejects a malformed (non-base64) digest", func(t *testing.T) {
		assert.Error(t, token.Validate("not-valid-base64!!!"))
	})

	t.Run("rejects a well-formed base64 digest that is not a bcrypt hash", func(t *testing.T) {
		assert.Error(t, token.Validate("bm90LWEtYmNyeXB0LWhhc2g="))
	})
}

func TestRandomToken(t *testing.T) {
	token, err := RandomToken()
	require.NoError(t, err)
	assert.Len(t, token.String(), 64) // sha256 hex digest

	other, err := RandomToken()
	require.NoError(t, err)
	assert.NotEqual(t, token.String(), other.String())
}

func TestRandomString(t *testing.T) {
	value, err := RandomString(32)
	require.NoError(t, err)
	assert.Len(t, value, 64) // sha256 hex digest, regardless of input size
}

func TestRandomBytes(t *testing.T) {
	value, err := RandomBytes(16)
	require.NoError(t, err)
	assert.Len(t, value, 16)

	other, err := RandomBytes(16)
	require.NoError(t, err)
	assert.NotEqual(t, value, other)
}

func TestNonce(t *testing.T) {
	a, err := Nonce()
	require.NoError(t, err)
	assert.NotEmpty(t, a)

	b, err := Nonce()
	require.NoError(t, err)
	assert.NotEqual(t, a, b)
}
