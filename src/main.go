package main

import (
	"github.com/NikosGour/BigDownloadP2P/src/build"
	log "github.com/NikosGour/logging/src"
)

func main(){
	log.Debug("DEBUG_MODE = %t\n",build.DEBUG_MODE)
}