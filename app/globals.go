package app

import (
	"fmt"
	"math"
	"os"
	"path"
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

	TEMP_B_SIZE = 256 * KiB
)

var (
	FILE_BUFFER_SIZE = int(4 * MiB) //300 * TEMP_B_SIZE

	PROJECT_DIR, _ = os.Getwd()
	DOWNLOADS_DIR  = path.Join(PROJECT_DIR, "downloads")
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

func BestUnitOfData(data int) (float32, string) {
	l := math.Log10(float64(data))
	switch {
	case l < 3:
		return float32(data), "B"
	case l >= 3 && l < 6:
		return float32(data) / float32(1000), "Kb"
	case l >= 6 && l < 9:
		return float32(data) / float32(1e6), "Mb"
	case l >= 9:
		return float32(data) / float32(1e9), "Gb"
	default:
		panic(fmt.Sprintf("Unreachable: %d, log10: %f", data, l))
	}
}
