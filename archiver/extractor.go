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
	"path/filepath"
)

func ExtractArchive(password, saltHex, archiveFile, outputDir string) error {
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return err
	}

	key, err := GenerateKey(password, salt)
	if err != nil {
		return err
	}

	inFile, err := os.Open(archiveFile)
	if err != nil {
		return err
	}
	defer inFile.Close()

	totalFiles, err := ReadHeader(inFile)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	for i := uint32(0); i < totalFiles; i++ {
		var nameLen uint16
		err := binary.Read(inFile, binary.BigEndian, &nameLen)
		if err != nil {
			return err
		}

		nameBytes := make([]byte, nameLen)
		if _, err := io.ReadFull(inFile, nameBytes); err != nil {
			return err
		}
		filename := string(nameBytes)

		var compressionMethod uint8
		if err := binary.Read(inFile, binary.BigEndian, &compressionMethod); err != nil {
			return err
		}

		var encryptedDataLen uint32
		if err := binary.Read(inFile, binary.BigEndian, &encryptedDataLen); err != nil {
			return err
		}

		encryptedData := make([]byte, encryptedDataLen)
		if _, err := io.ReadFull(inFile, encryptedData); err != nil {
			return err
		}

		decryptedData, err := Decrypt(encryptedData, key)
		if err != nil {
			return err
		}

		var originalData []byte
		switch compressionMethod {
		case CompressionNone:
			originalData = decryptedData
		case CompressionDeflate:
			originalData, err = Decompress(decryptedData)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown compression method: %d", compressionMethod)
		}
		outputPath := filepath.Join(outputDir, filename)

		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %v", filename, err)
		}

		if err := os.WriteFile(outputPath, originalData, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %v", outputPath, err)
		}
	}

	return nil
}

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
