package singletons

import (
	"log"

	"github.com/terabiome/terabiome/internal/config"
	"github.com/terabiome/terabiome/pkg/logging"
)

type Singletons struct {
}

var sings *Singletons

func Init() {
	log.Println("Constructing singletons")

	logging.InitGlobalSlogger(config.Get().Logger)

	sings = &Singletons{}
	// showing off

	logging.GetGlobalSlogger().Info("Constructed singletons")
}

func Get() *Singletons {
	return sings
}
