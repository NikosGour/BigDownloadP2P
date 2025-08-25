package app

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strconv"
	"time"

	log "github.com/NikosGour/logging/src"
	cmap "github.com/orcaman/concurrent-map/v2"
)

var (
	ErrUnrecognizedRequestType = errors.New("Unrecognized request type")
)

type Conn struct {
	c io.ReadWriteCloser
}

func NewConn(conn io.ReadWriteCloser) *Conn {
	c := &Conn{c: conn}
	return c
}

func (conn *Conn) Close() error {
	return conn.c.Close()
}

func (conn *Conn) Read(p []byte) (n int, err error) {
	return conn.c.Read(p)
}

func (conn *Conn) Write(p []byte) (n int, err error) {
	return conn.c.Write(p)
}

type Listener struct {
	Port                int
	DownloadsDir        string
	activeFileDownloads cmap.ConcurrentMap[UUID, *ActiveFileDownload]
}

type ActiveFileDownload struct {
	FileNames     []string
	DirName       string
	FileParts     int
	PartsFinished int
	DoneChan      chan int
	Done          bool
}

func NewActiveFileDownload(file_parts int) *ActiveFileDownload {
	afd := &ActiveFileDownload{FileParts: file_parts, Done: false}
	afd.FileNames = []string{}
	afd.DoneChan = make(chan int)
	go func() {
		for {
			if afd.FileParts <= afd.PartsFinished {
				afd.Done = true
				close(afd.DoneChan)
				return
			}
			<-afd.DoneChan
			afd.PartsFinished++
		}
	}()

	return afd
}

func NewListener(port int, downloads_dir string) *Listener {
	l := &Listener{Port: port}
	l.DownloadsDir = path.Join(PROJECT_DIR, "downloads")
	l.activeFileDownloads = cmap.NewStringer[UUID, *ActiveFileDownload]()

	if downloads_dir != "" {
		l.DownloadsDir = downloads_dir
	}

	return l
}

func (l *Listener) Listen() error {

	address := "0.0.0.0:" + strconv.Itoa(l.Port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("On listen: %w", err)
	}
	log.Info("Listening on `%s`", address)
	go l.handleActiveFileDownloads()

	for {
		conn, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("On accept: %w", err)
		}

		go l.handleConnection(NewConn(conn))
	}
}

func (l *Listener) handleActiveFileDownloads() {
	for {
		if l.activeFileDownloads.Count() != 0 {
			for tuple := range l.activeFileDownloads.IterBuffered() {
				active_file := tuple.Val
				if active_file.Done {
					l.FinalizeFileDownload(active_file.DirName)
				}
			}
		}
	}
}
func (l *Listener) FinalizeFileDownload(path string) {

}

func (l *Listener) handleConnection(conn *Conn) {
	defer conn.Close()
	timer := time.Now()

	request_header, err := conn.receiveRequestHeader()
	if err != nil {
		log.Error("%s", err)
		return
	}

	log.Debug("request_header=%s", request_header)

	switch request_header.RequestType {
	case RequestSendString:
		err = conn.receiveString()
		if err != nil {
			log.Error("%s", fmt.Errorf("On receive string: %w", err))
		}
	case RequestSendFile:
		err = conn.receiveFile(l, request_header)
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

func (conn *Conn) receiveRequestHeader() (RequestHeader, error) {
	json_request_header, err := conn.receiveSmallBytes()
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

func (conn *Conn) receiveString() error {
	buf, err := conn.readBytes()
	if err != nil {
		return err
	}

	log.Info("Message: `%s`", buf.String())
	return nil
}

func (conn *Conn) readBytes() (bytes.Buffer, error) {
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

func (conn *Conn) receiveSmallBytes() (bytes.Buffer, error) {
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

func receiveJson[T any](conn *Conn) (T, error) {
	var rv T
	data, err := conn.receiveSmallBytes()
	if err != nil {
		return rv, err
	}

	err = json.Unmarshal(data.Bytes(), &rv)
	if err != nil {
		return rv, fmt.Errorf("On unmarshal: %w", err)
	}

	return rv, nil
}

func (conn *Conn) receiveFile(l *Listener, request_header RequestHeader) error {
	// Create downloads dir if it doesn't exist
	downloads_dir, err := tryMakeNewDir(l.DownloadsDir)
	if err != nil {
		return err
	}
	l.DownloadsDir = downloads_dir

	// Get the file info
	file_info, err := receiveJson[FileInfoJSON](conn)
	if err != nil {
		return err
	}
	log.Debug("file_info=%#v", file_info)

	var file_dir string
	active_file, ok := l.activeFileDownloads.Get(request_header.UUID)
	if !ok {
		active_file = NewActiveFileDownload(NUMBER_OF_PARTS)
		l.activeFileDownloads.Set(request_header.UUID, active_file)

		file_dir, err = tryMakeNewDir(path.Join(l.DownloadsDir, file_info.Name))
		if err != nil {
			return err
		}

		active_file.DirName = file_dir
	}
	file_dir = active_file.DirName

	// Create the output file to add the content
	file_name := path.Join(file_dir, file_info.PartName)
	log.Debug("file_name=%#v", file_name)
	file, err := os.Create(file_name)
	if err != nil {
		return fmt.Errorf("On file create: %w", err)
	}
	defer file.Close()
	active_file.FileNames = append(active_file.FileNames, file_name)

	// Download the file using buffering
	bufferedWriter := bufio.NewWriterSize(file, FILE_BUFFER_SIZE)
	defer bufferedWriter.Flush()

	ticker := time.NewTicker(3 * time.Second)
	acc_bytes := 0
	total_bytes := 0
	buf := make([]byte, TEMP_B_SIZE)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			// log.Info("Read %d bytes from %s", n, conn.RemoteAddr())
			acc_bytes += n
			total_bytes += n
			// if total_bytes >= int(file_info.Size) {
			// 	n -= total_bytes - int(file_info.Size)
			// }
			_, writeErr := bufferedWriter.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("write failed: %w", writeErr)
			}

			// if total_bytes >= int(file_info.Size) {
			// 	return nil
			// }

			select {
			case <-ticker.C:
				transformed, unit := BestUnitOfData(acc_bytes / 3)
				log.Info("Download Speed: %f %s/sec", transformed, unit)
				acc_bytes = 0
			default:
			}
			reportDownloadProgress(file_info, total_bytes)

			// // Periodically flush to reduce memory use
			// if bufferedWriter.Buffered() > FILE_BUFFER_SIZE/2 {
			// 	_ = bufferedWriter.Flush()
			// }

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

func reportDownloadProgress(file_info FileInfoJSON, bytes_downloaded int) {
	// log.Info("Downloaded `%d/%d` %f%%", temp_file.Size(), file_info.Size, float64(temp_file.Size())/float64(file_info.Size))
}
