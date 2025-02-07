package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sethvargo/go-envconfig"
)

type Crypto interface {
	Encrypt(raw []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

type Reader interface {
	Read(cfg interface{}) error
}

type ReaderFunc func(cfg interface{}) error

func (cf ReaderFunc) Read(cfg interface{}) error {
	return cf(cfg)
}

func FromEnv(prefix string) ReaderFunc {
	lookuper := envconfig.PrefixLookuper(prefix, UpcaseLookuper(envconfig.OsLookuper()))

	return func(cfg interface{}) error {
		ecfg := &envconfig.Config{
			Target:   cfg,
			Lookuper: lookuper,
		}

		if err := envconfig.ProcessWith(context.Background(), ecfg); err != nil {
			return fmt.Errorf("failed to read configuration from environment: %w", err)
		}

		return nil
	}
}

func FromFile(filename string, optional bool) ReaderFunc {
	return func(cfg interface{}) error {
		configFileInfo, err := os.Stat(filename)

		if err != nil {
			if os.IsNotExist(err) && optional {
				return nil
			}

			return fmt.Errorf("failed to read configuration file %s: %w", filename, err)
		}

		if configFileInfo.IsDir() {
			return fmt.Errorf("configuration file %s can not be directory", filename)
		}

		data, err := os.ReadFile(filename)

		if err != nil {
			return fmt.Errorf("failed to read configuration file %s: %w", filename, err)
		}

		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("failed to parse configuration file %s: %w", filename, err)
		}

		return nil
	}
}

func FromEncryptedFile(filename string, optional bool, decryptor Crypto) ReaderFunc {
	return func(cfg interface{}) error {
		configFileInfo, err := os.Stat(filename)

		if err != nil {
			if os.IsNotExist(err) && optional {
				return nil
			}

			return fmt.Errorf("failed to read configuration file %s: %w", filename, err)
		}

		if configFileInfo.IsDir() {
			return fmt.Errorf("configuration file %s can not be directory", filename)
		}

		rawData, err := os.ReadFile(filename)

		if err != nil {
			return fmt.Errorf("failed to read configuration file %s: %w", filename, err)
		}

		data, err := decryptor.Decrypt(rawData)

		if err != nil {
			return fmt.Errorf("failed to decrypt configuration file %s: %w", filename, err)
		}

		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("failed to parse configuration file %s: %w", filename, err)
		}

		return nil
	}
}
