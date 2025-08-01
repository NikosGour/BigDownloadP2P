package app

import (
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

	n, err := io.CopyBuffer(file, conn, TEMP_B)
	log.Debug("n=%v", n)
	if err != nil {
		log.Error("%s", fmt.Errorf("On read: %w", err))
		return
	}
	// bufferedWriter := bufio.NewWriterSize(file, FILE_BUFFER_SIZE)
	// defer bufferedWriter.Flush()

	// for {
	// 	log.Debug("%p", &TEMP_B)
	// 	n, err := conn.Read(TEMP_B)
	// 	runtime.GC()
	// 	log.Debug("GC ran, read %d bytes", n)
	// 	if n > 0 {
	// 		_, writeErr := bufferedWriter.Write(TEMP_B[:n])
	// 		if writeErr != nil {
	// 			log.Error("write failed: %s", writeErr)
	// 			return
	// 		}

	// 		// Periodically flush to reduce memory use
	// 		if bufferedWriter.Buffered() > FILE_BUFFER_SIZE/2 {
	// 			_ = bufferedWriter.Flush()
	// 		}
	// 	}
	// 	if err == io.EOF {
	// 		_ = bufferedWriter.Flush()
	// 		break
	// 	}
	// 	if err != nil {
	// 		log.Error("read failed: %s", err)
	// 		return
	// 	}
	// }
}
