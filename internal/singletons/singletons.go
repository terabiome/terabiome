package singletons

import (
	"log"

	"github.com/terabiome/infrastructures/internal/config"
	"github.com/terabiome/infrastructures/pkg/logging"
)

type Singletons struct {
}

var sings *Singletons

func Init() {
	log.Println("Constructing singletons")

	logging.InitGlobalLogger(config.Get().Logger)

	sings = &Singletons{}
	// showing off

	logging.Get().Info("Constructed singletons")
}

func Get() *Singletons {
	return sings
}
