package app

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	log "github.com/NikosGour/logging/src"
)

type Sender struct {
	port int
	addr string
}

func NewFileSender(port int, address string) *Sender {
	fs := &Sender{port: port}
	fs.addr = address + ":" + strconv.Itoa(fs.port)
	return fs
}

func (fs *Sender) Send(data io.Reader, size int) error {
	//TODO: validate address
	conn, err := net.Dial("tcp", fs.addr)
	if err != nil {
		return fmt.Errorf("On dial: %w", err)
	}
	log.Debug("Connected on address: `%s`", fs.addr)

	err = binary.Write(conn, binary.BigEndian, int64(size))
	if err != nil {
		return fmt.Errorf("On write size: %w", err)
	}

	n, err := io.CopyN(conn, data, int64(size))
	log.Debug("n=%#v", n)
	if err != nil {
		return fmt.Errorf("On write: %w", err)
	}

	return nil
}

func (fs *Sender) SendString(data string) error {
	data_buffer := bytes.NewBufferString(data)
	err := fs.Send(data_buffer, data_buffer.Len())
	if err != nil {
		return err
	}

	return nil
}

func (fs *Sender) SendFile(file_path string) error {
	file, err := os.Open(file_path)
	if err != nil {
		return fmt.Errorf("On open: %w", err)
	}

	file_info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("On file.Stat(): %w", err)
	}

	err = fs.Send(file, int(file_info.Size()))
	if err != nil {
		return err
	}

	return nil
}

func (fs *Sender) SendFiles(file_paths []string) error {

	for i, file_path := range file_paths {
		err := fs.SendFile(file_path)
		if err != nil {
			return fmt.Errorf("On file number=`%d`, file_path=`%s` : %w", i, file_path, err)
		}

		log.Debug("Succesfully sent file: `%s`", file_path)
	}
	return nil
}
