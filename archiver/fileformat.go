package archiver

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	MagicNumber        = 0x53454146 // 'SEAF' in ASCII
	Version            = 1
	CompressionDeflate = 1 // Code for DEFLATE
	// I will add other compression methods in the future
)

// WriteHeader records the archive header, including the total number of files
func WriteHeader(w io.Writer, totalFiles uint32) error {
	// Writing a magic number
	if err := binary.Write(w, binary.BigEndian, uint32(MagicNumber)); err != nil {
		return err
	}
	// Writing a version
	if err := binary.Write(w, binary.BigEndian, uint16(Version)); err != nil {
		return err
	}
	// Writing the total number of files
	if err := binary.Write(w, binary.BigEndian, totalFiles); err != nil {
		return err
	}
	return nil
}

// ReadHeader reads the archive header and returns the total number of files
func ReadHeader(r io.Reader) (uint32, error) {
	var magic uint32
	if err := binary.Read(r, binary.BigEndian, &magic); err != nil {
		return 0, err
	}
	if magic != MagicNumber {
		return 0, errors.New("invalid archive format")
	}

	var version uint16
	if err := binary.Read(r, binary.BigEndian, &version); err != nil {
		return 0, err
	}
	if version != Version {
		return 0, fmt.Errorf("unsupported archive version: %d", version)
	}

	var totalFiles uint32
	if err := binary.Read(r, binary.BigEndian, &totalFiles); err != nil {
		return 0, err
	}

	return totalFiles, nil
}

// WriteFileEntry writes information about the file to the archive
func WriteFileEntry(w io.Writer, filename string, compressedData []byte, compressionMethod uint8, encryptedData []byte) error {
	// Writing the length of the file name
	nameLen := uint16(len(filename))
	if err := binary.Write(w, binary.BigEndian, nameLen); err != nil {
		return err
	}

	// Writing the file name
	if _, err := w.Write([]byte(filename)); err != nil {
		return err
	}

	// Writing a compression method
	if err := binary.Write(w, binary.BigEndian, compressionMethod); err != nil {
		return err
	}

	// Writing the length of the compressed data
	compressedDataLen := uint32(len(compressedData))
	if err := binary.Write(w, binary.BigEndian, compressedDataLen); err != nil {
		return err
	}

	// Writing compressed data
	if _, err := w.Write(compressedData); err != nil {
		return err
	}

	// Writing the length of the encrypted data
	encryptedDataLen := uint32(len(encryptedData))
	if err := binary.Write(w, binary.BigEndian, encryptedDataLen); err != nil {
		return err
	}

	// Writing encrypted data
	if _, err := w.Write(encryptedData); err != nil {
		return err
	}

	return nil
}
