package app

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

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

func (fs *Sender) Send(data *bufio.Reader) error {
	//TODO: validate address
	log.Debug("Dialing: %s", fs.addr)
	conn, err := net.Dial("tcp", fs.addr)
	if err != nil {
		return fmt.Errorf("On dial: %w", err)
	}
	log.Debug("Connected on address: `%s`", fs.addr)

	defer conn.Close()

	// err = binary.Write(conn, binary.BigEndian, int64(size))
	// if err != nil {
	// 	return fmt.Errorf("On write size: %w", err)
	// }

	n, err := io.CopyBuffer(conn, data, TEMP_B)
	log.Debug("n=%#v", n)
	if err != nil {
		return fmt.Errorf("On write: %w", err)
	}

	return nil
}

func (fs *Sender) SendString(data string) error {
	// data_reader := bytes.NewBufferString(data)
	data_reader := bufio.NewReaderSize(strings.NewReader(data), FILE_BUFFER_SIZE)
	err := fs.Send(data_reader)
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

	// file_info, err := file.Stat()
	// if err != nil {
	// 	return fmt.Errorf("On file.Stat(): %w", err)
	// }

	err = fs.Send(bufio.NewReaderSize(file, FILE_BUFFER_SIZE))
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
