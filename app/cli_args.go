package app

import (
	"errors"
	"flag"
	"fmt"
)

var (
	ErrAppCommandLineArgsNoFilesProvided = errors.New("No files were provided through arguments")
)

func commandLineArgs() (int, bool, []string, error) {
	const usage = `Usage: BigDownloadP2P [OPTIONS] [FILES]
Options:
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

	files := flag.Args()
	if !is_receiver && len(files) == 0 {
		return 0, false, nil, ErrAppCommandLineArgsNoFilesProvided
	}

	return port, is_receiver, files, nil
}
