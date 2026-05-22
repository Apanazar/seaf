package archiver

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"codeberg.org/tsukinoko-kun/oxipng-go"
	"github.com/gen2brain/jpegxl"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
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

func CreateArchive(password, saltHex, outputFile string, files []FileInfo, compressLevel int, optimizeImages bool, imageQuality float32) error {
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

	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, runtime.NumCPU())

	for _, file := range files {
		wg.Add(1)
		go func(f FileInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			data, err := os.ReadFile(f.Path)
			if err != nil {
				fmt.Printf("Error reading %s: %v\n", f.Path, err)
				return
			}

			if optimizeImages {
				optData, changed, optErr := OptimizeImage(data, f.Path, imageQuality)
				if optErr != nil {
					fmt.Printf("Warning: could not optimize %s: %v\n", f.Path, optErr)
				} else if changed {
					fmt.Printf("Optimized %s: %d bytes -> %d bytes\n",
						filepath.Base(f.Path), len(data), len(optData))
					data = optData
				}
			}

			dataToStore, method, err := PrepareEntryData(data, compressLevel)
			if err != nil {
				fmt.Printf("Error preparing %s: %v\n", f.Path, err)
				return
			}

			encryptedData, err := Encrypt(dataToStore, key)
			if err != nil {
				fmt.Printf("Error encrypting %s: %v\n", f.Path, err)
				return
			}

			mu.Lock()
			err = WriteFileEntry(outFile, filepath.Base(f.Path), method, encryptedData)
			mu.Unlock()
			if err != nil {
				fmt.Printf("Error writing entry %s: %v\n", f.Path, err)
			}
		}(file)
	}

	wg.Wait()

	return nil
}

func OptimizeImage(originalData []byte, filename string, imageQuality float32) ([]byte, bool, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	if ext == ".png" {
		optimized, err := oxipng.Optimize(originalData)
		if err != nil {
			fmt.Printf("Warning: oxipng failed for %s: %v\n", filename, err)
			return originalData, false, nil
		}
		if len(optimized) < len(originalData) {
			return optimized, true, nil
		}
		return originalData, false, nil
	}

	img, _, err := image.Decode(bytes.NewReader(originalData))
	if err != nil {
		return originalData, false, nil
	}

	var buf bytes.Buffer
	quality := 75
	if imageQuality >= 1.0 {
		quality = 100
	} else if imageQuality > 0 {
		quality = int(imageQuality * 100)
		if quality < 1 {
			quality = 1
		} else if quality > 99 {
			quality = 99
		}
	}
	opts := jpegxl.Options{
		Quality: quality,
		Effort:  3,
	}

	if err := jpegxl.Encode(&buf, img, opts); err != nil {
		fmt.Printf("Warning: JPEG XL encoding failed for %s: %v\n", filename, err)
		return originalData, false, nil
	}

	jxlData := buf.Bytes()
	if len(jxlData) < len(originalData) {
		return jxlData, true, nil
	}
	return originalData, false, nil
}

func hasTransparency(img image.Image) bool {
	switch src := img.(type) {
	case *image.NRGBA:
		for i := 3; i < len(src.Pix); i += 4 {
			if src.Pix[i] < 255 {
				return true
			}
		}
	case *image.RGBA:
		for i := 3; i < len(src.Pix); i += 4 {
			if src.Pix[i] < 255 {
				return true
			}
		}
	case *image.Gray:
		return false
	case *image.Paletted:
		for _, c := range src.Palette {
			if _, _, _, a := c.RGBA(); a < 0xffff {
				return true
			}
		}
		return false
	default:
		return false
	}
	return false
}

func PrepareEntryData(data []byte, compressLevel int) ([]byte, uint8, error) {
	compressed, err := Compress(data, compressLevel)
	if err != nil {
		return nil, 0, err
	}

	if len(compressed) < len(data) {
		return compressed, CompressionDeflate, nil
	}
	return data, CompressionNone, nil
}
