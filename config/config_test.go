package config

import (
	"log/slog"
	"testing"

	"github.com/sethvargo/go-envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppEnvIs(t *testing.T) {
	assert.True(t, Production.Is(Production))
	assert.False(t, AppEnv("development").Is(Production))
}

func TestUpcaseLookuper(t *testing.T) {
	next := LookuperFunc(func(key string) (string, bool) {
		if key == "CRYPTO_KEY" {
			return "value", true
		}

		return "", false
	})

	lookuper := UpcaseLookuper(next)

	t.Run("upcases the key before delegating", func(t *testing.T) {
		value, ok := lookuper.Lookup("crypto_key")
		assert.True(t, ok)
		assert.Equal(t, "value", value)
	})

	t.Run("reports misses from the wrapped lookuper", func(t *testing.T) {
		_, ok := lookuper.Lookup("unknown")
		assert.False(t, ok)
	})
}

func TestLookuperFuncSatisfiesEnvconfig(t *testing.T) {
	var lookuper envconfig.Lookuper = LookuperFunc(func(string) (string, bool) {
		return "", false
	})

	_, ok := lookuper.Lookup("anything")
	assert.False(t, ok)
}

func TestNewLogger(t *testing.T) {
	t.Run("logs at info and above in production", func(t *testing.T) {
		logger := NewLogger(BaseConfig{Env: Production})
		require.NotNil(t, logger)

		assert.True(t, logger.Enabled(t.Context(), slog.LevelInfo))
		assert.False(t, logger.Enabled(t.Context(), slog.LevelDebug))
	})

	t.Run("logs at debug and above outside production", func(t *testing.T) {
		logger := NewLogger(BaseConfig{Env: "development"})
		require.NotNil(t, logger)

		assert.True(t, logger.Enabled(t.Context(), slog.LevelDebug))
	})
}
