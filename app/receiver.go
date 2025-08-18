package app

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"os"
	"path"
	"strconv"
	"time"

	log "github.com/NikosGour/logging/src"
)

var (
	ErrUnrecognizedRequestType = errors.New("Unrecognized request type")
)

func Listen(port int, downloads_dir string) error {

	if downloads_dir != "" {
		DOWNLOADS_DIR = downloads_dir
	}

	address := "0.0.0.0:" + strconv.Itoa(port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("On listen: %w", err)
	}
	log.Info("Listening on `%s`", address)

	for {
		conn, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("On accept: %w", err)
		}
		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	timer := time.Now()

	request_header, err := receiveRequestHeader(conn)
	if err != nil {
		log.Error("%s", err)
		return
	}

	log.Debug("request_header=%s", request_header)

	switch request_header.RequestType {
	case RequestSendString:
		err = receiveString(conn)
		if err != nil {
			log.Error("%s", fmt.Errorf("On receive string: %w", err))
		}
	case RequestSendFile:
		err = receiveFile(conn)
		if err != nil {
			log.Error("%s", fmt.Errorf("On receive file: %w", err))
		}
	default:
		err = ErrUnrecognizedRequestType
		if err != nil {
			log.Error("%s", fmt.Errorf("%w: %d", err, request_header.RequestType))
		}
	}

	log.Info("Download took %s", time.Since(timer))
}

func receiveRequestHeader(conn net.Conn) (RequestHeader, error) {
	json_request_header, err := receiveSmallBytes(conn)
	if err != nil {
		return RequestHeader{}, err
	}

	var request_header RequestHeader
	err = json.Unmarshal(json_request_header.Bytes(), &request_header)
	if err != nil {
		return RequestHeader{}, fmt.Errorf("On Unmarshal json: %w", err)
	}

	return request_header, nil
}

func receiveString(conn net.Conn) error {
	buf, err := readBytes(conn)
	if err != nil {
		return err
	}

	log.Info("Message: `%s`", buf.String())
	return nil
}

func readBytes(conn net.Conn) (bytes.Buffer, error) {
	buf := make([]byte, TEMP_B_SIZE)
	data := bytes.Buffer{}
	n, err := io.CopyBuffer(&data, conn, buf)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("On read: %w", err)
	}
	if n <= 0 {
		log.Warn("Read `%d` bytes for object", n)
	}

	return data, nil
}

func receiveSmallBytes(conn net.Conn) (bytes.Buffer, error) {
	var size int64
	err := binary.Read(conn, binary.BigEndian, &size)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("On read size:%w", err)
	}

	data := bytes.Buffer{}
	n, err := io.CopyN(&data, conn, size)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("On read object: %w", err)
	}

	if n <= 0 {
		log.Warn("Read `%d` bytes for object", n)
	}
	return data, nil
}

func receiveJson[T any](conn net.Conn) (T, error) {
	var rv T
	data, err := receiveSmallBytes(conn)
	if err != nil {
		return rv, err
	}

	err = json.Unmarshal(data.Bytes(), &rv)
	if err != nil {
		return rv, fmt.Errorf("On unmarshal: %w", err)
	}

	return rv, nil
}

func receiveFile(conn net.Conn) error {
	// Get the file info
	file_info, err := receiveJson[FileInfoJSON](conn)
	if err != nil {
		return err
	}
	log.Debug("file_info=%#v", file_info)

	// Create output dir if it doesn't exist
	for i := 0; ; i++ {
		var downloads_dir string
		if i == 0 {
			downloads_dir = DOWNLOADS_DIR
		} else {
			downloads_dir = DOWNLOADS_DIR + "_" + strconv.Itoa(i)
		}
		err = os.Mkdir(downloads_dir, os.ModeDir|os.ModePerm)

		if os.IsExist(err) {
			temp_dir_info, err := os.Stat(downloads_dir)
			if err != nil {
				return fmt.Errorf("On stat: %w", err)
			}

			if !temp_dir_info.Mode().IsDir() {
				continue
			}

		} else if err != nil {
			return fmt.Errorf("On Mkdir: %w", err)
		}
		DOWNLOADS_DIR = downloads_dir
		break
	}

	// Create the output file to add the content
	file_name := path.Join(DOWNLOADS_DIR, strconv.Itoa(int(time.Now().Unix()))+"_"+strconv.Itoa(int(rand.Int32()))+"_"+file_info.Name)
	log.Debug("file_name=%#v", file_name)
	file, err := os.Create(file_name)
	if err != nil {
		return fmt.Errorf("On file create: %w", err)
	}
	defer file.Close()

	// Download the file using buffering
	bufferedWriter := bufio.NewWriterSize(file, FILE_BUFFER_SIZE)
	defer bufferedWriter.Flush()

	ticker := time.NewTicker(3 * time.Second)
	acc_byte := 0

	buf := make([]byte, TEMP_B_SIZE)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			// log.Info("Read %d bytes from %s", n, conn.RemoteAddr())
			acc_byte += n

			_, writeErr := bufferedWriter.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("write failed: %w", writeErr)
			}
			temp_info, err := os.Stat(file_name)
			if err != nil {
				return fmt.Errorf("On temp Stat: %w", err)
			}
			select {
			case <-ticker.C:
				transformed, unit := BestUnitOfData(acc_byte / 3)
				log.Info("Download Speed: %f %s/sec", transformed, unit)
				acc_byte = 0
			default:
			}
			reportDownloadProgress(file_info, temp_info)

			// Periodically flush to reduce memory use
			if bufferedWriter.Buffered() > FILE_BUFFER_SIZE/2 {
				_ = bufferedWriter.Flush()
			}

		}
		if err == io.EOF {
			_ = bufferedWriter.Flush()
			break
		}
		if err != nil {
			return fmt.Errorf("read failed: %w", err)
		}
	}
	return nil
}

func reportDownloadProgress(file_info FileInfoJSON, temp_file os.FileInfo) {
	// log.Info("Downloaded `%d/%d` %f%%", temp_file.Size(), file_info.Size, float64(temp_file.Size())/float64(file_info.Size))
}
