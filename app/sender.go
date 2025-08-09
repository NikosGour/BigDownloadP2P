package app

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"

	log "github.com/NikosGour/logging/src"
)

type UUID = uuid.UUID

type Sender struct {
	port int
	addr string
	conn net.Conn
}

func NewFileSender(port int, address string) *Sender {
	fs := &Sender{port: port}
	fs.addr = address + ":" + strconv.Itoa(fs.port)

	conn, err := fs.connect()
	if err != nil {
		log.Fatal("%s", err)
	}
	fs.conn = conn

	return fs
}

func (fs *Sender) connect() (net.Conn, error) {
	//TODO: validate address
	log.Info("Dialing: %s", fs.addr)
	conn, err := net.Dial("tcp", fs.addr)
	if err != nil {
		return nil, fmt.Errorf("On dial: %w", err)
	}
	log.Info("Connected on address: `%s`", fs.addr)
	return conn, nil
}

func (fs *Sender) sendBytes(data io.Reader, request_header RequestHeader) error {
	err := fs.requestPrologue(request_header)
	if err != nil {
		return err
	}

	buf := make([]byte, TEMP_B_SIZE)
	n, err := io.CopyBuffer(fs.conn, data, buf)
	log.Debug("n=%#v", n)
	if err != nil {
		return fmt.Errorf("On write: %w", err)
	}

	return nil
}

func (fs *Sender) requestPrologue(request_header RequestHeader) error {
	err := fs.sendRequestHeader(request_header)
	if err != nil {
		return err
	}
	return nil
}

func (fs *Sender) sendHandlePackets(data io.Reader, request_header RequestHeader, packetHandling func(n int)) error {
	err := fs.requestPrologue(request_header)
	if err != nil {
		return err
	}

	return fs.sendHandlePacketsNoRequestHeader(data, packetHandling)
}

func (fs *Sender) sendHandlePacketsNoRequestHeader(data io.Reader, packetHandling func(n int)) error {
	buf := make([]byte, TEMP_B_SIZE)
	for {
		_n, err := data.Read(buf)
		if _n > 0 {
			n, err := fs.conn.Write(buf[:_n])
			if err != nil {
				return fmt.Errorf("write failed: %w", err)
			}

			if n <= 0 {
				log.Warn("Wrote `%d` bytes", n)
			}

			packetHandling(n)

		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read failed: %w", err)
		}
	}

}
func (fs *Sender) sendRequestHeader(request_header RequestHeader) error {
	id_json, err := json.Marshal(request_header)
	if err != nil {
		return fmt.Errorf("On marshal header: %w", err)
	}

	n, err := fs.sendSmallBytes(id_json)
	if err != nil {
		return err
	}
	if n <= 0 {
		log.Warn("Wrote `%d` bytes", n)
	}
	return nil
}

func (fs *Sender) sendSmallBytes(data []byte) (int64, error) {
	err := binary.Write(fs.conn, binary.BigEndian, int64(len(data)))
	if err != nil {
		return 0, fmt.Errorf("On write data size: %w", err)
	}

	n, err := io.CopyN(fs.conn, bytes.NewBuffer(data), int64(len(data)))
	if err != nil {
		return 0, fmt.Errorf("On send data body: %w", err)
	}
	return n, nil
}

func (fs *Sender) SendJson(data any, request_header RequestHeader) (int64, error) {
	err := fs.requestPrologue(request_header)
	if err != nil {
		return 0, err
	}
	return fs.sendJsonNoHeader(data)
}

func (fs *Sender) sendJsonNoHeader(data any) (int64, error) {
	data_json, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("On marshal: %w", err)
	}

	n, err := fs.sendSmallBytes(data_json)
	if err != nil {
		return 0, err
	}

	if n <= 0 {
		log.Warn("Wrote `%d` bytes", n)
	}

	return n, nil
}

func (fs *Sender) SendString(data string) error {
	data_reader := bufio.NewReaderSize(strings.NewReader(data), FILE_BUFFER_SIZE)
	rh := RequestHeader{UUID: uuid.New(), RequestType: RequestSendString}
	log.Debug("request_header=%s", rh)

	err := fs.sendBytes(data_reader, rh)
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

	rh := RequestHeader{UUID: uuid.New(), RequestType: RequestSendFile}
	log.Debug("request_header=%s", rh)

	err = fs.sendRequestHeader(rh)
	if err != nil {
		return err
	}

	file_info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("On file.Stat(): %w", err)
	}

	n, err := fs.sendJsonNoHeader(FromFileInfo(file_info))
	if err != nil {
		return err
	}
	if n <= 0 {
		log.Warn("Wrote `%d` bytes", n)
	}

	err = fs.sendHandlePacketsNoRequestHeader(bufio.NewReaderSize(file, FILE_BUFFER_SIZE), func(n int) {
		log.Info("Wrote %d bytes into %s", n, fs.conn.RemoteAddr())
	})
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

func (fs *Sender) Close() error {
	return fs.conn.Close()
}
