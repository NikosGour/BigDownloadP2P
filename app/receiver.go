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
	"time"

	log "github.com/NikosGour/logging/src"
)

func Receive(port int) error {
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
		timer := time.Now()
		handleConnection(conn)
		log.Debug("Download took %s", time.Since(timer))
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	file, err := os.Create("temp_file")
	if err != nil {
		log.Error("%s", fmt.Errorf("On file create: %w", err))
		return
	}
	defer file.Close()

	uuid, err := receiveRequestUUID(conn)
	if err != nil {
		log.Error("%s", err)
		return
	}
	log.Info("uuid=%s", uuid)

	err = receiveLoop(file, conn)
	if err != nil {
		log.Error("%s", fmt.Errorf("On receive loop: %w", err))
	}
}

func receiveRequestUUID(conn net.Conn) (UUID, error) {
	json_uuid, err := receiveSmallBytes(conn)
	if err != nil {
		return UUID{}, err
	}

	var uuid UUID
	err = json.Unmarshal(json_uuid.Bytes(), &uuid)
	if err != nil {
		return UUID{}, fmt.Errorf("On Unmarshal json: %w", err)
	}

	return uuid, nil
}

func receiveSmallBytes(conn net.Conn) (bytes.Buffer, error) {
	var size int64
	err := binary.Read(conn, binary.BigEndian, &size)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("On read json_size: %w", err)
	}

	data := bytes.Buffer{}
	n, err := io.CopyN(&data, conn, size)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("On read json: %w", err)
	}

	if n <= 0 {
		log.Warn("Read `%d` bytes for json", n)
	}
	return data, nil
}

func receiveLoop(file *os.File, conn net.Conn) error {
	bufferedWriter := bufio.NewWriterSize(file, FILE_BUFFER_SIZE)
	defer bufferedWriter.Flush()

	buf := make([]byte, TEMP_B_SIZE)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			log.Info("Read %d bytes from %s", n, conn.RemoteAddr())
			_, writeErr := bufferedWriter.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("write failed: %w", writeErr)
			}

			reportDownloadProgress(file)

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

func reportDownloadProgress(file *os.File) {

}
