package app

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"

	log "github.com/NikosGour/logging/src"
)

func Receive(port int) error {
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

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	var size int64
	content := bytes.Buffer{}

	err := binary.Read(conn, binary.BigEndian, &size)
	if err != nil {
		log.Error("%s", fmt.Errorf("On read size: %w", err))
		return
	}

	n, err := io.CopyN(&content, conn, size)
	log.Debug("n=%v", n)
	if err != nil {
		log.Error("%s", fmt.Errorf("On read: %w", err))
		return
	}

	log.Debug("Content: `%s`", content.String())
}
