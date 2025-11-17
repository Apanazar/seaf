package archiver

import (
	"bytes"
	"compress/flate"
	"fmt"
	"io"
)

func Compress(data []byte, level int) ([]byte, error) {
	if level < 0 || level > 9 {
		return nil, fmt.Errorf("invalid compression level: %d, must be between 0 and 9", level)
	}

	var buf bytes.Buffer
	writer, err := flate.NewWriter(&buf, level)
	if err != nil {
		return nil, err
	}

	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func Decompress(data []byte) ([]byte, error) {
	reader := flate.NewReader(bytes.NewReader(data))
	defer reader.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
