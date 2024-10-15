package archiver

import (
	"bytes"
	"compress/flate"
	"io"
)

// Compress compresses data using the DEFLATE algorithm
func Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	// Creating a new flat.Writer with compression level 5 (default)
	writer, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}

	//Writing data to flate.Writer
	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		return nil, err
	}

	// Closing the writer to write down all the remaining data
	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Decompress decompresses data using the DEFLATE algorithm
func Decompress(data []byte) ([]byte, error) {
	reader := flate.NewReader(bytes.NewReader(data))
	defer reader.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
