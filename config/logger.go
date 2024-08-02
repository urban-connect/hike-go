package config

import "go.uber.org/zap"

func NewLogger(cfg *BaseConfig) (*zap.SugaredLogger, error) {
	var logger *zap.Logger
	var err error

	if cfg.Env.Is(Production) {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment(
			zap.IncreaseLevel(zap.DebugLevel),
		)
	}

	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}
