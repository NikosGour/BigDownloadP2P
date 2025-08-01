package app

const (
	Kibibyte int64 = 1024
	KiB            = Kibibyte
	Mebibyte       = Kibibyte * 1024
	MiB            = Mebibyte
	Gibibyte       = Mebibyte * 1024
	GiB            = Gibibyte
	Tebibyte       = Gibibyte * 1024
	TiB            = Tebibyte
)

var (
	TEMP_B           = make([]byte, 1*MiB)
	FILE_BUFFER_SIZE = 300 * len(TEMP_B)
)
