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
	"time"

	"github.com/google/uuid"
	cmap "github.com/orcaman/concurrent-map/v2"

	log "github.com/NikosGour/logging/src"
)

type UUID = uuid.UUID

type Sender struct {
	port  int
	addr  string
	conns cmap.ConcurrentMap[UUID, net.Conn]
}

func NewFileSender(port int, address string) *Sender {
	fs := &Sender{port: port}
	fs.addr = address + ":" + strconv.Itoa(fs.port)

	fs.conns = cmap.NewStringer[UUID, net.Conn]()

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

	err = conn.(*net.TCPConn).SetNoDelay(true)
	if err != nil {
		return nil, fmt.Errorf("On Nagle's: %w", err)
	}
	return conn, nil
}

func (conn *Conn) sendBytes(data io.Reader, request_header RequestHeader) error {
	err := conn.requestPrologue(request_header)
	if err != nil {
		return err
	}

	buf := make([]byte, TEMP_B_SIZE)
	n, err := io.CopyBuffer(conn, data, buf)
	log.Debug("n=%#v", n)
	if err != nil {
		return fmt.Errorf("On write: %w", err)
	}

	return nil
}

func (conn *Conn) requestPrologue(request_header RequestHeader) error {
	err := conn.sendRequestHeader(request_header)
	if err != nil {
		return err
	}
	return nil
}

func (conn *Conn) sendHandlePackets(data io.Reader, request_header RequestHeader, packetHandling func(n int)) error {
	err := conn.requestPrologue(request_header)
	if err != nil {
		return err
	}
	//TODO fix
	return conn.sendHandlePacketsNoRequestHeader(data, 0, packetHandling)
}

func (conn *Conn) sendHandlePacketsNoRequestHeader(data io.Reader, count int, packetHandling func(n int)) error {
	ticker := time.NewTicker(3 * time.Second)
	acc_byte := 0
	bytes_read := 0
	buf := make([]byte, TEMP_B_SIZE)
	for {
		_n, err := data.Read(buf)
		if _n > 0 {
			bytes_read += _n
			if bytes_read > count {
				_n -= bytes_read - count
			}
			n, err := conn.Write(buf[:_n])
			acc_byte += n
			if err != nil {
				return fmt.Errorf("write failed: %w", err)
			}

			// if n <= 0 {
			// 	log.Warn("Wrote `%d` bytes", n)
			// }

			select {
			case <-ticker.C:
				transformed, unit := BestUnitOfData(acc_byte / 3)
				log.Info("Upload Speed: %f %s/sec", transformed, unit)
				acc_byte = 0
			default:
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
func (conn *Conn) sendRequestHeader(request_header RequestHeader) error {
	id_json, err := json.Marshal(request_header)
	if err != nil {
		return fmt.Errorf("On marshal header: %w", err)
	}

	n, err := conn.sendSmallBytes(id_json)
	if err != nil {
		return err
	}
	if n <= 0 {
		log.Warn("Wrote `%d` bytes", n)
	}
	return nil
}

func (conn *Conn) sendSmallBytes(data []byte) (int64, error) {
	err := binary.Write(conn, binary.BigEndian, int64(len(data)))
	if err != nil {
		return 0, fmt.Errorf("On write data size: %w", err)
	}

	n, err := io.CopyN(conn, bytes.NewBuffer(data), int64(len(data)))
	if err != nil {
		return 0, fmt.Errorf("On send data body: %w", err)
	}
	return n, nil
}

func (conn *Conn) SendJson(data any, request_header RequestHeader) (int64, error) {
	err := conn.requestPrologue(request_header)
	if err != nil {
		return 0, err
	}
	return conn.sendJsonNoHeader(data)
}

func (conn *Conn) sendJsonNoHeader(data any) (int64, error) {
	data_json, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("On marshal: %w", err)
	}

	n, err := conn.sendSmallBytes(data_json)
	if err != nil {
		return 0, err
	}

	if n <= 0 {
		log.Warn("Wrote `%d` bytes", n)
	}

	return n, nil
}

func (conn *Conn) SendString(data string) error {
	data_reader := bufio.NewReaderSize(strings.NewReader(data), FILE_BUFFER_SIZE)
	rh := RequestHeader{UUID: uuid.New(), RequestType: RequestSendString}
	log.Debug("request_header=%s", rh)

	err := conn.sendBytes(data_reader, rh)
	if err != nil {
		return err
	}

	return nil
}

func (fs *Sender) SendFile(file_path string) error {

	readers, part_size, err := fs.splitFileIntoParts(file_path)
	if err != nil {
		return err
	}
	file_info, err := os.Stat(file_path)
	if err != nil {
		return fmt.Errorf("On Stat: %w", err)
	}

	// Todo:  Split file in part and send
	uuid := uuid.New()
	for i, reader := range readers {
		_conn, err := fs.connect()
		if err != nil {
			return err
		}
		conn := NewConn(_conn)

		log.Debug("Sending part %d", i)
		err = conn.sendFilePart(reader, part_size, file_info, i, uuid)
		if err != nil {
			return err
		}
	}

	return nil
}

func (fs *Sender) splitFileIntoParts(file_path string) ([]*bufio.Reader, int64, error) {
	file, err := os.Open(file_path)
	if err != nil {
		return nil, 0, fmt.Errorf("On open: %w", err)
	}
	file_info, err := file.Stat()
	if err != nil {
		return nil, 0, fmt.Errorf("On file.Stat(): %w", err)
	}

	files := []*os.File{}

	for i := 0; i < NUMBER_OF_PARTS; i++ {
		file, err := os.Open(file_path)
		if err != nil {
			return nil, 0, fmt.Errorf("On open: %w", err)
		}
		files = append(files, file)
	}

	file_parts := [NUMBER_OF_PARTS]*bufio.Reader{}
	part_size := int64(file_info.Size() / 4)

	for i := 0; i < NUMBER_OF_PARTS; i += 1 {
		seek_pos := part_size * int64(i)
		_, err := files[i].Seek(seek_pos, 0)
		if err != nil {
			return nil, 0, fmt.Errorf("On seek: %w", err)
		}
		file_parts[i] = bufio.NewReaderSize(files[i], FILE_BUFFER_SIZE)
	}

	// for i, v := range file_parts {
	// 	buf, _, err := v.ReadLine()
	// 	if err != nil && err != io.EOF {
	// 		return nil, 0, fmt.Errorf("On readLine: %w", err)
	// 	}
	// 	log.Debug("buf%d=%s", i, string(buf[:NUMBER_OF_PARTS]))

	// }
	return file_parts[:], part_size, nil
}
func (conn *Conn) sendFilePart(r *bufio.Reader, n int64, file_info os.FileInfo, part_num int, uuid UUID) error {
	rh := RequestHeader{UUID: uuid, RequestType: RequestSendFile}
	log.Debug("request_header=%s", rh)

	err := conn.sendRequestHeader(rh)
	if err != nil {
		return err
	}
	file_info_json := FromFileInfo(file_info)
	file_info_json.Size = n
	file_info_json.PartName = file_info_json.Name + strconv.Itoa(part_num)

	_n, err := conn.sendJsonNoHeader(file_info_json)
	if err != nil {
		return err
	}
	if _n <= 0 {
		log.Warn("Wrote `%d` bytes", _n)
	}

	err = conn.sendHandlePacketsNoRequestHeader(r, int(n), func(n int) {
		// log.Info("Wrote %d bytes into %s", n, fs.conn.RemoteAddr())
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
	//TODO: implement
	return nil
}
