package logging

import (
	"github.com/sirupsen/logrus"
	"os"
	"pdm-logic-server/pkg/config"
)

func NewLogger(cfg *config.LogConfig) (*logrus.Logger, error) {
	logger := logrus.New()

	// Set output format
	if cfg.JSON {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	// Set output
	if cfg.File != "" {
		file, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		logger.SetOutput(file)
	}

	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	logger.SetLevel(level)

	return logger, nil
}
