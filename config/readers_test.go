package config

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/urban-connect/hike-go/crypto"
)

type testConfig struct {
	BaseConfig

	API struct {
		Host string `json:"host" env:"api_host"`
		Port int    `json:"port" env:"api_port"`
	} `json:"api"`
}

func writeFile(t *testing.T, dir, name string, data []byte) string {
	t.Helper()

	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, data, 0o600))

	return path
}

func newTestAES(t *testing.T) *crypto.AES {
	t.Helper()

	aes, err := crypto.NewAES(
		base64.StdEncoding.EncodeToString(make([]byte, 32)),
		base64.StdEncoding.EncodeToString(make([]byte, 12)),
	)
	require.NoError(t, err)

	return aes
}

func TestFromEnv(t *testing.T) {
	t.Run("reads values through the prefixed, upcased lookuper", func(t *testing.T) {
		t.Setenv("HIKE_VERSION", "1.2.3")
		t.Setenv("HIKE_API_HOST", "localhost")
		t.Setenv("HIKE_API_PORT", "8001")

		var cfg testConfig
		require.NoError(t, FromEnv("HIKE_").Read(&cfg))

		assert.Equal(t, "1.2.3", cfg.Version)
		assert.Equal(t, "localhost", cfg.API.Host)
		assert.Equal(t, 8001, cfg.API.Port)
	})

	t.Run("leaves fields untouched when no env vars are set", func(t *testing.T) {
		var cfg testConfig
		require.NoError(t, FromEnv("MISSING_").Read(&cfg))

		assert.Empty(t, cfg.Version)
		assert.Empty(t, cfg.API.Host)
	})
}

func TestFromFile(t *testing.T) {
	dir := t.TempDir()

	t.Run("reads JSON into the config", func(t *testing.T) {
		path := writeFile(t, dir, "settings.json", []byte(`{"version":"1.0.0","api":{"host":"example.com","port":443}}`))

		var cfg testConfig
		require.NoError(t, FromFile(path, false).Read(&cfg))

		assert.Equal(t, "1.0.0", cfg.Version)
		assert.Equal(t, "example.com", cfg.API.Host)
		assert.Equal(t, 443, cfg.API.Port)
	})

	t.Run("skips a missing file when optional", func(t *testing.T) {
		var cfg testConfig
		assert.NoError(t, FromFile(filepath.Join(dir, "nope.json"), true).Read(&cfg))
	})

	t.Run("fails on a missing file when required", func(t *testing.T) {
		var cfg testConfig
		assert.Error(t, FromFile(filepath.Join(dir, "nope.json"), false).Read(&cfg))
	})

	t.Run("fails when the path is a directory", func(t *testing.T) {
		var cfg testConfig
		assert.Error(t, FromFile(dir, true).Read(&cfg))
	})

	t.Run("fails on malformed JSON", func(t *testing.T) {
		path := writeFile(t, dir, "broken.json", []byte(`{"version":`))

		var cfg testConfig
		assert.Error(t, FromFile(path, false).Read(&cfg))
	})
}

func TestFromEncryptedFile(t *testing.T) {
	dir := t.TempDir()
	aes := newTestAES(t)

	encrypted, err := aes.Encrypt([]byte(`{"version":"2.0.0","api":{"host":"secret.example.com"}}`))
	require.NoError(t, err)

	t.Run("decrypts and reads JSON into the config", func(t *testing.T) {
		path := writeFile(t, dir, "settings.json.encrypted", encrypted)

		var cfg testConfig
		require.NoError(t, FromEncryptedFile(path, false, aes).Read(&cfg))

		assert.Equal(t, "2.0.0", cfg.Version)
		assert.Equal(t, "secret.example.com", cfg.API.Host)
	})

	t.Run("skips a missing file when optional", func(t *testing.T) {
		var cfg testConfig
		assert.NoError(t, FromEncryptedFile(filepath.Join(dir, "nope.encrypted"), true, aes).Read(&cfg))
	})

	t.Run("fails when the payload cannot be decrypted", func(t *testing.T) {
		path := writeFile(t, dir, "garbage.encrypted", []byte("not actually encrypted"))

		var cfg testConfig
		assert.Error(t, FromEncryptedFile(path, false, aes).Read(&cfg))
	})
}

// The layering contract: readers run in order and each one overrides the last.
func TestReadersLayerInOrder(t *testing.T) {
	dir := t.TempDir()

	base := writeFile(t, dir, "base.json", []byte(`{"version":"base","api":{"host":"base.example.com","port":80}}`))
	local := writeFile(t, dir, "local.json", []byte(`{"version":"local"}`))

	t.Setenv("HIKE_API_PORT", "9999")

	var cfg testConfig
	readers := []Reader{
		FromEnv("HIKE_"),
		FromFile(base, false),
		FromFile(local, true),
	}

	for _, reader := range readers {
		require.NoError(t, reader.Read(&cfg))
	}

	// local.json wins over base.json for version...
	assert.Equal(t, "local", cfg.Version)
	// ...base.json still supplies what local.json omits...
	assert.Equal(t, "base.example.com", cfg.API.Host)
	// ...and the later file readers override the earlier env reader.
	assert.Equal(t, 80, cfg.API.Port)
}
