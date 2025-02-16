package decompress

import (
	"compress/gzip"
	"io"
)

func Gzip(data io.ReadCloser) ([]byte, error) {
	gzipReader, err := gzip.NewReader(data)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	return io.ReadAll(gzipReader)
}
