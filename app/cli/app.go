package cli

import (
	"github.com/NikosGour/BigDownloadP2P/app"
	log "github.com/NikosGour/logging/src"
)

func Start() {
	port, is_receiver, address, files, err := commandLineArgs()
	if err != nil {
		log.Fatal("%s", err)
	}

	log.Debug("files=%v", files)

	if is_receiver {
		err = app.Receive(port)
	} else {
		fs := app.NewFileSender(port, address)
		err = fs.SendFiles(files)
	}

	if err != nil {
		log.Fatal("%s", err)
	}
}
