package app

import (
	"flag"
	"fmt"
)

func Start() {
	port, is_receiver := commandLineArgs()
	if is_receiver {
		receiver(port)
	} else {
		sender(port)
	}
}

func commandLineArgs() (int, bool) {
	const usage = `Usage of BigDownloadP2P:
		-p | --port		Define the port for the client to use (default: 6969)
		-r | --is_receiver	Toggle if the client is a sender or a receiver (default: sender)`

	var port int
	flag.IntVar(&port, "p", 6969, "The port for the client to use")
	flag.IntVar(&port, "port", 6969, "The port for the client to use")

	var is_receiver bool
	flag.BoolVar(&is_receiver, "r", false, "Will toggle the client to receive instead of send files")
	flag.BoolVar(&is_receiver, "is_receiver", false, "Will toggle the client to receive instead of send files")

	flag.Usage = func() { fmt.Println(usage) }
	flag.Parse()

	return port, is_receiver
}

func receiver(port int) {
	//TODO
}

func sender(port int) {
	//TODO
}
