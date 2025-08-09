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

	if is_receiver {
		err = app.Receive(port)
	} else {
		log.Debug("files=%v", files)
		fs := app.NewFileSender(port, address)
		// err = fs.SendFiles(files)
		err = fs.SendString("nikos")
	}

	if err != nil {
		log.Fatal("%s", err)
	}
}
