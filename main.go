package main

import (
	"github.com/NikosGour/BigDownloadP2P/app"
	"github.com/NikosGour/BigDownloadP2P/build"
	log "github.com/NikosGour/logging/src"
)

func main() {
	log.Debug("DEBUG_MODE = %t", build.DEBUG_MODE)
	app.Main()
}
