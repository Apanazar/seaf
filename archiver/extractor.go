package archiver

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"

	"seaf/ui"
)

// ExtractArchive extracts files from an archive
func ExtractArchive(password, saltHex, archiveFile string) error {
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

	// Opening the archive file
	inFile, err := os.Open(archiveFile)
	if err != nil {
		return err
	}
	defer inFile.Close()

	// Reading the title and the total number of files
	totalFiles, err := ReadHeader(inFile)
	if err != nil {
		return err
	}

	// Initialize the progress bar with the total number of files
	progressUI := ui.NewProgressUI(int(totalFiles))

	// Reading file entries
	for i := uint32(0); i < totalFiles; i++ {
		// Reading the length of the file name
		var nameLen uint16
		err := binary.Read(inFile, binary.BigEndian, &nameLen)
		if err != nil {
			return err
		}

		// Reading the file name
		nameBytes := make([]byte, nameLen)
		if _, err := io.ReadFull(inFile, nameBytes); err != nil {
			return err
		}
		filename := string(nameBytes)

		// Reading the compression method
		var compressionMethod uint8
		if err := binary.Read(inFile, binary.BigEndian, &compressionMethod); err != nil {
			return err
		}

		// Reading the length of the compressed data
		var compressedDataLen uint32
		if err := binary.Read(inFile, binary.BigEndian, &compressedDataLen); err != nil {
			return err
		}

		// Reading compressed data
		compressedData := make([]byte, compressedDataLen)
		if _, err := io.ReadFull(inFile, compressedData); err != nil {
			return err
		}

		// Reading the length of the encrypted data
		var encryptedDataLen uint32
		if err := binary.Read(inFile, binary.BigEndian, &encryptedDataLen); err != nil {
			return err
		}

		// Reading encrypted data
		encryptedData := make([]byte, encryptedDataLen)
		if _, err := io.ReadFull(inFile, encryptedData); err != nil {
			return err
		}

		// Decrypting the data
		decryptedData, err := Decrypt(encryptedData, key)
		if err != nil {
			return err
		}

		// Unpacking the data
		var originalData []byte
		switch compressionMethod {
		case CompressionDeflate:
			originalData, err = Decompress(decryptedData)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown compression method: %d", compressionMethod)
		}

		// Writing the file to disk
		if err := os.WriteFile(filename, originalData, 0644); err != nil {
			return err
		}

		// Updating the progress bar
		progressUI.Increment()
	}

	progressUI.Finish()
	return nil
}

// Decrypt decrypts data using AES-GCM
func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("incorrect ciphertext")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
