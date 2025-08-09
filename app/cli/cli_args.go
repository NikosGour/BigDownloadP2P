package cli

import (
	"errors"
	"flag"
	"fmt"
)

var (
	ErrAppCommandLineArgsNoFilesProvided = errors.New("No files were provided through arguments")
)

func commandLineArgs() (port int, is_receiver bool, address string, output_dir string, files []string, err error) {
	const usage = `Usage: BigDownloadP2P [OPTIONS] [FILES]
Files:
	You can pass space seperated file paths at the end of the command to send to the address. E.g. BigDownloadP2P -p 4444 ./a.txt ./b.log ./c.exe
Options:
		-p | --port		Define the port for the client to use (default: 6969)
		-r | --is_receiver	Toggle if the client is a sender or a receiver (default: sender)
		-a | --address	The destination ip address (default: localhost)
		-o | --output_dir The dir where or the downloads will be placed (default: pwd)
		`

	flag.IntVar(&port, "p", 6969, "The port for the client to use")
	flag.IntVar(&port, "port", 6969, "The port for the client to use")

	flag.BoolVar(&is_receiver, "r", false, "Will toggle the client to receive instead of send files")
	flag.BoolVar(&is_receiver, "is_receiver", false, "Will toggle the client to receive instead of send files")

	flag.StringVar(&address, "a", "localhost", "The ip address to send the files to")
	flag.StringVar(&address, "address", "localhost", "The ip address to send the files to")

	flag.StringVar(&output_dir, "o", "", "The output directory to place the downloads")
	flag.StringVar(&output_dir, "output_dir", "", "The output directory to place the downloads")

	flag.Usage = func() { fmt.Println(usage) }
	flag.Parse()

	files = flag.Args()
	if !is_receiver && len(files) == 0 {
		err = ErrAppCommandLineArgsNoFilesProvided
		return
	}

	return
}
