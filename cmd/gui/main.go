package main

import (
	"github.com/NikosGour/BigDownloadP2P/app/gui"
	"github.com/NikosGour/BigDownloadP2P/build"
	log "github.com/NikosGour/logging/src"
)

func main() {
	if build.DEBUG_MODE {
		log.Debug("DEBUG MODE")
	} else {
		log.Debug("RELEASE MODE")
	}

	gui.Start()
}
