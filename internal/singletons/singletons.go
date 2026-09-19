package singletons

import (
	"github.com/terabiome/infrastructures/internal/config"
	"github.com/terabiome/infrastructures/pkg/logger"
)

type Singletons struct {
	logger *logger.Logger
}

var sings *Singletons

func Init() {
	loggerInstance := logger.New(config.Get().Logger)
	sings = &Singletons{
		logger: loggerInstance,
	}
}

func (s *Singletons) Logger() *logger.Logger {
	return s.logger
}

func Get() *Singletons {
	return sings
}
