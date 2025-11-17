package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"

	"seaf/archiver"
	"seaf/ui"
)

var (
	password      string
	saltHex       string
	outputFile    string
	extract       bool
	archiveFile   string
	generateSalt  bool
	saltLength    int
	compressLevel int
)

func main() {
	if len(os.Args) == 1 {
		runGUI()
	} else {
		runTUI()
	}
}

func runTUI() {
	flag.Parse()

	if generateSalt {
		var err error
		saltHex, err = generateRandomSalt(saltLength)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating salt: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated salt (hex): %s\n", saltHex)
	}
	if password == "" || saltHex == "" {
		fmt.Println("You must specify the password and the salt.")
		flag.Usage()
		os.Exit(1)
	}

	if extract {
		err := archiver.ExtractArchive(password, saltHex, archiveFile)
		if err != nil {
			log.Fatalf("Error extracting the archive: %v", err)
		}

		fmt.Println("The extraction was completed successfully!")
	} else {
		inputFiles := flag.Args()
		if len(inputFiles) == 0 {
			fmt.Println("No input files are specified.")
			flag.Usage()
			os.Exit(1)
		}

		files, err := archiver.CollectFiles(inputFiles)
		if err != nil {
			log.Fatalf("Error when collecting files: %v", err)
		}

		totalOriginalSize := int64(0)
		totalCompressedSize := int64(0)
		totalEncryptedSize := int64(0)

		salt, err := hex.DecodeString(saltHex)
		if err != nil {
			log.Fatalf("Error decoding salt: %v", err)
		}

		key, err := archiver.GenerateKey(password, salt)
		if err != nil {
			log.Fatalf("Error generating key: %v", err)
		}

		for _, file := range files {
			data, err := os.ReadFile(file.Path)
			if err != nil {
				log.Fatalf("Error reading file %s: %v", file.Path, err)
			}

			compressedData, err := archiver.Compress(data, compressLevel)
			if err != nil {
				log.Fatalf("Error compressing file %s: %v", file.Path, err)
			}

			encryptedData, err := archiver.Encrypt(compressedData, key)
			if err != nil {
				log.Fatalf("Error encrypting file %s: %v", file.Path, err)
			}

			fmt.Printf("\n--- Processing: %s ---\n", file.Path)
			fmt.Printf("Original size: %d bytes (%.2f MB)\n", len(data), float64(len(data))/(1024*1024))

			entropy := calculateEntropy(data)
			fmt.Printf("Data entropy: %.4f (0-1, the higher it is, the worse it shrinks)\n", entropy)

			compressionRatio := float64(len(compressedData)) / float64(len(data)) * 100
			fmt.Printf("After compression: %d bytes, Ratio: %.2f%%\n", len(compressedData), compressionRatio)
			fmt.Printf("After encryption: %d bytes\n", len(encryptedData))
			fmt.Printf("Total overhead: %+d bytes\n", len(encryptedData)-len(data))

			totalOriginalSize += int64(len(data))
			totalCompressedSize += int64(len(compressedData))
			totalEncryptedSize += int64(len(encryptedData))
		}

		fmt.Printf("\n=== FINAL RESULTS ===\n")
		fmt.Printf("Original total: %d bytes (%.2f MB)\n", totalOriginalSize, float64(totalOriginalSize)/(1024*1024))
		fmt.Printf("Compressed total: %d bytes (%.2f MB)\n", totalCompressedSize, float64(totalCompressedSize)/(1024*1024))
		fmt.Printf("Encrypted total: %d bytes (%.2f MB)\n", totalEncryptedSize, float64(totalEncryptedSize)/(1024*1024))
		fmt.Printf("Final compression: %.2f%%\n", float64(totalCompressedSize)/float64(totalOriginalSize)*100)
		fmt.Printf("Archive overhead: %.2f%%\n", float64(totalEncryptedSize-totalOriginalSize)/float64(totalOriginalSize)*100)

		if err := createOutputDir(); err != nil {
			log.Fatalf("Error creating output directory: %v", err)
		}

		fullOutputPath := filepath.Join("output", outputFile)

		err = archiver.CreateArchive(password, saltHex, fullOutputPath, files, compressLevel)
		if err != nil {
			log.Fatalf("Error creating the archive: %v", err)
		}

		fmt.Printf("Archive successfully created: %s\n", fullOutputPath)
		fmt.Println("Archiving and encryption have been completed successfully.")
	}
}

func createOutputDir() error {
	if _, err := os.Stat("output"); os.IsNotExist(err) {
		fmt.Println("Creating output directory...")
		err := os.Mkdir("output", 0755)
		if err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
		fmt.Println("Output directory created successfully")
	} else if err != nil {
		return fmt.Errorf("error checking output directory: %v", err)
	}
	return nil
}

func runGUI() {
	gui := ui.NewGUI()
	gui.ShowAndRun()
}

func calculateEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0
	}

	var freq [256]int
	for _, b := range data {
		freq[b]++
	}

	var entropy float64
	for _, count := range freq {
		if count > 0 {
			p := float64(count) / float64(len(data))
			entropy -= p * math.Log2(p)
		}
	}

	return entropy / 8.0
}

func init() {
	flag.StringVar(&password, "password", "", "Password for encryption/decryption")
	flag.StringVar(&saltHex, "salt", "", "Salt (in hexadecimal format)")
	flag.StringVar(&outputFile, "output", "archive.seaf", "The name of the output file to archive")
	flag.BoolVar(&extract, "extract", false, "Extract files from the archive")
	flag.StringVar(&archiveFile, "archive", "archive.seaf", "The name of the archive to extract")
	flag.BoolVar(&generateSalt, "generate-salt", false, "Generate a random salt")
	flag.IntVar(&saltLength, "salt-length", 16, "Length of the generated salt in bytes")
	flag.IntVar(&compressLevel, "compress", 6, "Compression level (0-9, where 0=no compression, 1=fastest, 9=best compression)")

	asciiArt := `
              _____                    _____                    _____                    _____          
             /\    \                  /\    \                  /\    \                  /\    \         
            /::\    \                /::\    \                /::\    \                /::\    \        
           /::::\    \              /::::\    \              /::::\    \              /::::\    \       
          /::::::\    \            /::::::\    \            /::::::\    \            /::::::\    \      
         /:::/\:::\    \          /:::/\:::\    \          /:::/\:::\    \          /:::/\:::\    \     
        /:::/__\:::\    \        /:::/__\:::\    \        /:::/__\:::\    \        /:::/__\:::\    \    
        \:::\   \:::\    \      /::::\   \:::\    \      /::::\   \:::\    \      /::::\   \:::\    \   
      ___\:::\   \:::\    \    /::::::\   \:::\    \    /::::::\   \:::\    \    /::::::\   \:::\    \  
     /\   \:::\   \:::\    \  /:::/\:::\   \:::\    \  /:::/\:::\   \:::\    \  /:::/\:::\   \:::\    \ 
    /::\   \:::\   \:::\____\/:::/__\:::\   \:::\____\/:::/  \:::\   \:::\____\/:::/  \:::\   \:::\____\
    \:::\   \:::\   \::/    /\:::\   \:::\   \::/    /\::/    \:::\  /:::/    /\::/    \:::\   \::/    /
     \:::\   \:::\   \/____/  \:::\   \:::\   \/____/  \/____/ \:::\/:::/    /  \/____/ \:::\   \/____/ 
      \:::\   \:::\    \       \:::\   \:::\    \               \::::::/    /            \:::\    \     
       \:::\   \:::\____\       \:::\   \:::\____\               \::::/    /              \:::\____\    
        \:::\  /:::/    /        \:::\   \::/    /               /:::/    /                \::/    /    
         \:::\/:::/    /          \:::\   \/____/               /:::/    /                  \/____/     
          \::::::/    /            \:::\    \                  /:::/    /                               
           \::::/    /              \:::\____\                /:::/    /                                
            \::/    /                \::/    /                \::/    /                                 
             \/____/                  \/____/                  \/____/                                  
                                                                                                        
	`

	flag.Usage = func() {
		fmt.Println("Usage of S.E.A.F.:")
		fmt.Println(asciiArt)
		fmt.Println("Secure Archiver is a tool for encrypting and archiving files.")
		fmt.Println()
		fmt.Println("Run without arguments to launch GUI interface")
		fmt.Println("Or use command line flags for terminal usage:")
		fmt.Println()
		fmt.Println("Flags:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  Archive files:")
		fmt.Println("    ", "./seaf", "--password=... --salt=... --output=archive.seaf file1 file2")
		fmt.Println("    Output: ./output/archive.seaf")
		fmt.Println()
		fmt.Println("  Generate salt and archive files:")
		fmt.Println("    ", "./seaf", "--password=... --generate-salt --salt-length=16 --output=archive.seaf file1 file2")
		fmt.Println("    Output: ./output/archive.seaf")
		fmt.Println()
		fmt.Println("  Extract files:")
		fmt.Println("    ", "./seaf", "--password=... --salt=... --extract --archive=archive.seaf")
		fmt.Println()
		fmt.Println("  Launch GUI:")
		fmt.Println("    ", "./seaf")
	}
}

func generateRandomSalt(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid salt length: %d", length)
	}
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(salt), nil
}
