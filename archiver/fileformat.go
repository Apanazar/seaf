package archiver

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	MagicNumber        = 0x53454146
	Version            = 1
	CompressionDeflate = 6
	// I will add other compression methods in the future
)

func WriteHeader(w io.Writer, totalFiles uint32) error {
	if err := binary.Write(w, binary.BigEndian, uint32(MagicNumber)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(Version)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, totalFiles); err != nil {
		return err
	}
	return nil
}

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

func WriteFileEntry(w io.Writer, filename string, compressionMethod uint8, encryptedData []byte) error {
	nameLen := uint16(len(filename))
	if err := binary.Write(w, binary.BigEndian, nameLen); err != nil {
		return err
	}

	if _, err := w.Write([]byte(filename)); err != nil {
		return err
	}

	if err := binary.Write(w, binary.BigEndian, compressionMethod); err != nil {
		return err
	}

	encryptedDataLen := uint32(len(encryptedData))
	if err := binary.Write(w, binary.BigEndian, encryptedDataLen); err != nil {
		return err
	}

	if _, err := w.Write(encryptedData); err != nil {
		return err
	}

	return nil
}
