package app

import (
	"fmt"
	"math"
	"os"
	"strconv"
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

	NUMBER_OF_PARTS = 4
)

var (
	FILE_BUFFER_SIZE = int(4 * MiB) //300 * TEMP_B_SIZE

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

	PartName string `json:"part_name"`
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

func tryMakeNewDir(path string) (string, error) {
	for i := 0; ; i++ {
		var downloads_dir string
		if i == 0 {
			downloads_dir = path
		} else {
			downloads_dir = path + "_" + strconv.Itoa(i)
		}
		err := os.Mkdir(downloads_dir, os.ModeDir|os.ModePerm)

		if os.IsExist(err) {
			temp_dir_info, err := os.Stat(downloads_dir)
			if err != nil {
				return "", fmt.Errorf("On stat: %w", err)
			}

			if !temp_dir_info.Mode().IsDir() {
				continue
			}

		} else if err != nil {
			return "", fmt.Errorf("On Mkdir: %w", err)
		}
		return downloads_dir, nil
	}
}
