package archiver

import (
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
)

type FileInfo struct {
	Path string
	Size int64
}

func CollectFiles(paths []string) ([]FileInfo, error) {
	var files []FileInfo
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			return nil, errors.New("directories are not supported")
		}
		files = append(files, FileInfo{Path: path, Size: info.Size()})
	}
	return files, nil
}

func CreateArchive(password, saltHex, outputFile string, files []FileInfo, compressLevel int) error {
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return err
	}

	key, err := GenerateKey(password, salt)
	if err != nil {
		return err
	}

	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	if err := WriteHeader(outFile, uint32(len(files))); err != nil {
		return err
	}

	for _, file := range files {
		data, err := os.ReadFile(file.Path)
		if err != nil {
			return err
		}

		compressedData, err := Compress(data, compressLevel)
		if err != nil {
			return err
		}

		encryptedData, err := Encrypt(compressedData, key)
		if err != nil {
			return err
		}

		if err := WriteFileEntry(outFile, filepath.Base(file.Path), CompressionDeflate, encryptedData); err != nil {
			return err
		}
	}

	return nil
}
