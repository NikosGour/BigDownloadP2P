package app

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"

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
		err = sender(port, address, files)
	}

	if err != nil {
		log.Fatal("%s", err)
	}
}

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

func receiver(port int) error {
	//TODO
	address := "0.0.0.0:" + strconv.Itoa(port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("On listen: %w", err)
	}
	log.Debug("Listening on `%s`", address)

	for {
		conn, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("On accept: %w", err)
		}

		go func() {
			var content []byte
			n, err := conn.Read(content)
			log.Debug("n=%#v", n)

			if err != nil {
				log.Error("%s", fmt.Errorf("On read: %w", err))
				return
			}

			log.Debug("Content: %s", string(content))
		}()
	}
}

func sender(port int, address string, file_names []string) error {
	//TODO: validate address
	address = address + ":" + strconv.Itoa(port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("On dial: %w", err)
	}
	log.Debug("Connected on address: `%s`", address)

	files := strings.Join(file_names, " ") + "\n"
	n, err := conn.Write([]byte(files))
	log.Debug("n=%#v", n)
	if err != nil {
		return fmt.Errorf("On write: %w", err)
	}

	return nil
}
