package main

import (
	// "net/http"
	// _ "net/http/pprof"

	"github.com/NikosGour/BigDownloadP2P/app/cli"
	"github.com/NikosGour/BigDownloadP2P/build"
	log "github.com/NikosGour/logging/src"
	// "github.com/pkg/profile"
)

func main() {
	// defer profile.Start(profile.MemProfile).Stop()
	// go func() {
	// 	http.ListenAndServe(":8080", nil)
	// }()

	if build.DEBUG_MODE {
		log.Debug("DEBUG MODE")
	} else {
		log.Debug("RELEASE MODE")
	}

	cli.Start()
}
