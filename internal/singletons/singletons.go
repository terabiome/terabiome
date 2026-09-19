package singletons

import (
	"log"

	"github.com/terabiome/infrastructures/internal/config"
	"github.com/terabiome/infrastructures/pkg/logger"
)

type Singletons struct {
}

var sings *Singletons

func Init() {
	log.Println("Constructing singletons")

	logger.InitGlobalLogger(config.Get().Logger)

	sings = &Singletons{}
	// showing off

	logger.Get().Info("Constructed singletons")
}

func Get() *Singletons {
	return sings
}
