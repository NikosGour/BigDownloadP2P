package app

import (
	"fmt"
	"os"
	"time"
)

const (
	Kibibyte int64 = 1024
	KiB            = Kibibyte
	Mebibyte       = Kibibyte * 1024
	MiB            = Mebibyte
	Gibibyte       = Mebibyte * 1024
	GiB            = Gibibyte
	Tebibyte       = Gibibyte * 1024
	TiB            = Tebibyte

	TEMP_B_SIZE = 1 * MiB
)

var (
	TEMP_B = make([]byte, TEMP_B_SIZE)

	FILE_BUFFER_SIZE = 300 * len(TEMP_B)

	PROJECT_DIR, _ = os.Getwd()
)

type RequestType int

const (
	RequestSendString RequestType = iota
	RequestSendFile
)

//go:generate easytags $GOFILE
type RequestHeader struct {
	RequestType RequestType `json:"request_type"`
	UUID        UUID        `json:"uuid"`
}

func (rh RequestHeader) String() string {
	return fmt.Sprintf("{UUID:%s, RequestType:%d}", rh.UUID, rh.RequestType)
}

type FileInfoJSON struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	Mode    string    `json:"mode"`
	ModTime time.Time `json:"mod_time"`
	IsDir   bool      `json:"is_dir"`
}

func FromFileInfo(info os.FileInfo) FileInfoJSON {
	return FileInfoJSON{
		Name:    info.Name(),
		Size:    info.Size(),
		Mode:    info.Mode().String(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	}

}
