package archiver

import (
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"

	"seaf/ui"
)

// FileInfo contains information about the file to archive
type FileInfo struct {
	Path string
	Size int64
}

// CollectFiles collects information about files
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

// CreateArchive creates an encrypted archive
func CreateArchive(password, saltHex, outputFile string, files []FileInfo, progressUI *ui.ProgressUI) error {
	// Decode the salt
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return err
	}

	// Generating a key
	key, err := GenerateKey(password, salt)
	if err != nil {
		return err
	}

	// Creating an output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Writing a header with the total number of files
	if err := WriteHeader(outFile, uint32(len(files))); err != nil {
		return err
	}

	// Processing files
	for _, file := range files {
		// Reading the contents of the file
		data, err := os.ReadFile(file.Path)
		if err != nil {
			return err
		}

		// Compressing the data
		compressedData, err := Compress(data)
		if err != nil {
			return err
		}

		// Encrypting compressed data
		encryptedData, err := Encrypt(compressedData, key)
		if err != nil {
			return err
		}

		// Writing a file entry to the archive
		if err := WriteFileEntry(outFile, filepath.Base(file.Path), compressedData, CompressionDeflate, encryptedData); err != nil {
			return err
		}

		// Updating the progress bar
		progressUI.Increment()
	}

	return nil
}
