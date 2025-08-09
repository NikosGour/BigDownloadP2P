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
	log.Debug("Listening on `%s`", address)

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

	var json_size int64
	err = binary.Read(conn, binary.BigEndian, &json_size)
	if err != nil {
		log.Error("%s", fmt.Errorf("On read json_size: %w", err))
		return
	}

	json_id := bytes.Buffer{}
	_n, err := io.CopyN(&json_id, conn, json_size)
	if err != nil {
		log.Error("%s", fmt.Errorf("On read json: %w", err))
		return
	}

	if _n <= 0 {
		log.Warn("Read `%d` bytes for json", _n)
	}

	var uuid UUID
	err = json.Unmarshal(json_id.Bytes(), &uuid)
	if err != nil {
		log.Error("%s", fmt.Errorf("On Unmarshal json: %w", err))
		return
	}

	log.Debug("uuid=%s", uuid)

	// n, err := io.CopyBuffer(file, conn, TEMP_B)
	// log.Debug("n=%v", n)
	// if err != nil {
	// 	log.Error("%s", fmt.Errorf("On read: %w", err))
	// 	return
	// }
	err = receiveLoop(file, conn)
	if err != nil {
		log.Error("%s", fmt.Errorf("On receive loop: %w", err))
	}
}

func receiveLoop(file *os.File, conn net.Conn) error {
	bufferedWriter := bufio.NewWriterSize(file, FILE_BUFFER_SIZE)
	defer bufferedWriter.Flush()

	for {
		n, err := conn.Read(TEMP_B)
		if n > 0 {
			_, writeErr := bufferedWriter.Write(TEMP_B[:n])
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
