package config

type BaseConfig struct {
	Env     AppEnv `json:"env,omitempty"`
	Version string `json:"version" env:"version"`

	Crypto struct {
		Key   string `json:"key,omitempty" env:"crypto_key"`
		Nonce string `json:"nonce,omitempty" env:"crypto_nonce"`
	} `json:"crypto,omitempty"`
}

type AppEnv string

const (
	Production AppEnv = "production"
)

func (env AppEnv) Is(other AppEnv) bool {
	return env == other
}
