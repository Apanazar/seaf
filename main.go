package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"

	"seaf/archiver"
	"seaf/ui"
)

var (
	password     string
	saltHex      string
	outputFile   string
	extract      bool
	archiveFile  string
	generateSalt bool
	saltLength   int
)

func main() {
	flag.Parse()

	// Checking and generating salt if the --generate-salt flag is set
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
		// Extract files from the archive with 3 attempts to complete the game
		err := extractArchiveWithGame(password, saltHex, archiveFile)
		if err != nil {
			log.Fatalf("Error extracting the archive: %v", err)
		}
		fmt.Println("The extraction was completed successfully!")
	} else {
		// File Archiving
		inputFiles := flag.Args()
		if len(inputFiles) == 0 {
			fmt.Println("No input files are specified.")
			flag.Usage()
			os.Exit(1)
		}

		// Collecting information about files
		files, err := archiver.CollectFiles(inputFiles)
		if err != nil {
			log.Fatalf("Error when collecting files: %v", err)
		}

		// Initializing the UI
		progressUI := ui.NewProgressUI(len(files))

		// Starting archiving and encryption
		err = archiver.CreateArchive(password, saltHex, outputFile, files, progressUI)
		if err != nil {
			log.Fatalf("Error creating the archive: %v", err)
		}

		progressUI.Finish()
		fmt.Println("Archiving and encryption have been completed successfully.")
	}
}

// Defining command line flags
func init() {
	flag.StringVar(&password, "password", "", "Password for encryption/decryption")
	flag.StringVar(&saltHex, "salt", "", "Salt (in hexadecimal format)")
	flag.StringVar(&outputFile, "output", "archive.seaf", "The name of the output file to archive")
	flag.BoolVar(&extract, "extract", false, "Extract files from the archive")
	flag.StringVar(&archiveFile, "archive", "archive.seaf", "The name of the archive to extract")
	flag.BoolVar(&generateSalt, "generate-salt", false, "Generate a random salt")
	flag.IntVar(&saltLength, "salt-length", 16, "Length of the generated salt in bytes")

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
		fmt.Println("Flags:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  Archive files:")
		fmt.Println("    ", "./seaf", "--password=... --salt=... --output=archive.seaf file1 file2")
		fmt.Println()
		fmt.Println("  Generate salt and archive files:")
		fmt.Println("    ", "./seaf", "--password=... --generate-salt --salt-length=16 --output=archive.seaf file1 file2")
		fmt.Println()
		fmt.Println("  Extract files:")
		fmt.Println("    ", "./seaf", "--password=... --salt=... --extract --archive=archive.seaf")
	}
}

// generateRandomSalt generates a random salt of a given length
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

// extractArchiveWithGame performs archive extraction after successful completion of the game
func extractArchiveWithGame(password, saltHex, archiveFile string) error {
	maxAttempts := 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		fmt.Printf("Attempt %d from %d: Complete the game to extract the archive.\n", attempt, maxAttempts)
		result := ui.RunGame()
		if result.Success {
			fmt.Println(result.Message)
			// Starting extraction
			err := archiver.ExtractArchive(password, saltHex, archiveFile)
			if err != nil {
				return err
			}
			return nil
		} else {
			fmt.Println("Try to turn off the computer, you won't escape anyway!")
		}
	}

	// If all attempts are unsuccessful, delete the archive
	fmt.Println("All attempts are unsuccessful. The archive will be deleted.")
	err := os.Remove(archiveFile)
	if err != nil {
		return fmt.Errorf("the archive could not be deleted: %v", err)
	}

	return fmt.Errorf("extraction failed after %d attempts", maxAttempts)
}
