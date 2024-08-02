package config

type Config struct {
	Env     string `json:"env,omitempty"`
	Version string `json:"version" env:"version"`

	Crypto struct {
		Key   string `json:"key,omitempty" env:"crypto_key"`
		Nonce string `json:"nonce,omitempty" env:"crypto_nonce"`
	} `json:"crypto,omitempty"`
}
