package app

import (
	log "github.com/NikosGour/logging/src"
)

func Start() {
	port, is_receiver, files, err := commandLineArgs()
	if err != nil {
		log.Fatal("%s", err)
	}

	log.Debug("files=%v", files)

	address := "localhost"
	if is_receiver {
		err = receiver(port)
	} else {
		fs := NewFileSender(port, address)
		err = fs.SendFiles(files)
	}

	if err != nil {
		log.Fatal("%s", err)
	}
}
