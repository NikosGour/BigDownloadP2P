package cli

import (
	"github.com/NikosGour/BigDownloadP2P/app"
	log "github.com/NikosGour/logging/src"
)

func Start() {

	port, is_receiver, address, output_dir, files, err := commandLineArgs()
	if err != nil {
		log.Fatal("%s", err)
	}

	if is_receiver {
		l := app.NewListener(port, output_dir)
		err = l.Listen()
	} else {
		log.Debug("files=%v", files)
		fs := app.NewFileSender(port, address)
		defer fs.Close()

		err = fs.SendFiles(files)
		// err = fs.SendString("nikos")
	}

	if err != nil {
		log.Fatal("%s", err)
	}
}
