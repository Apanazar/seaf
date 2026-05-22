# S.E.A.F. - Secure Encrypted Archive Format

```
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
                                                                                                    
```


## Overview

**S.E.A.F. (Secure Encrypted Archive Format)** is a robust and secure archiving tool developed in Go. It allows users to compress and encrypt multiple files into a single, secure archive with a custom format designed for enhanced security and scalability.

## Features

- **Encryption with AES-GCM**: Ensures data confidentiality and integrity using Advanced Encryption Standard in Galois/Counter Mode.
- **Compression with DEFLATE**: Efficiently compresses data to reduce archive size.
- **Custom Archive Format**: Unique `.seaf` format distinguishes your archives from standard formats, reducing vulnerability to known exploits.
- **Interactive Security Challenge**: Users must successfully complete a game with 3 attempts to extract files, adding an extra layer of security.
- **Cross-Platform Support**: Easily build binaries for Unix and Windows systems using Go's built-in cross-compilation features.
- **Salt Generation**: Automatically generate cryptographic salts with customizable lengths for enhanced security.
- **Optimizing image files**: The `oxipng-go` and `jpegxl` libraries optimize and compress images in the best possible way.
> `ATTENTION: Optimizing image files it takes a lot of time and a multi-core processor`

## Installation

### Building from Source

- **Clone the Repository**

```bash
git clone https://github.com/Apanazar/seaf.git
cd seaf
```

## For Unix Systems (Linux/macOS):
`GOOS=linux  GOARCH=amd64 go build --ldflags="-s -w" -o seaf-linux`  
`GOOS=darwin GOARCH=amd64 go build --ldflags="-s -w" -o seaf-macos-amd64`  
`GOOS=darwin GOARCH=arm64 go build --ldflags="-s -w" -o seaf-macos-arm64`  

## For Windows Systems:
`GOOS=windows GOARCH=amd64 go build --ldflags="-s -w -H windowsgui" -o seaf-windows.exe` for only GUI  
or  
`GOOS=windows GOARCH=amd64 go build --ldflags="-s -w" -o seaf-windows.exe` for GUI and TUI modes

## Usage

### Command-Line Flags

- `--password <str>`       Password for encryption/decryption
- `--salt <hex>`           Salt in hexadecimal format
- `--output <file>`        Output archive file name (default: archive.seaf)
- `--extract`              Extract files from archive
- `--archive <file>`       Archive file to extract (default: archive.seaf)
- `--generate-salt`        Generate a random salt
- `--salt-length <bytes>`  Length of generated salt in bytes (default: 16)
- `--compress <0-9>`       Compression level (0=none, 1=fastest, 9=best) (default: 6)
- `--optimize-images`      Lossless recompression of PNG, convert JPEG/other to JPEG XL
- `--quality <0-100>`      JPEG XL quality, 100 = lossless (default: 75)
- `--help`                 Display this help

### Archiving Files:
`./seaf --password=... --salt=... --output=archive.seaf file1 file2`

### Generating a Random Salt:
`./seaf --password=... --generate-salt --salt-length=16 --output=archive.seaf file1 file2`

### Extracting Files:
`./seaf --password=... --salt=... --extract --archive=archive.seaf`


## Security Advantages
1. AES-GCM Encryption: Utilizes a strong encryption standard ensuring data confidentiality and integrity.
2. Unique Archive Format: Custom .seaf format reduces susceptibility to vulnerabilities associated with common archive formats.
3. Salt Usage: Incorporates cryptographic salts to prevent rainbow table attacks and enhance password security.

## Contact
For any inquiries or support, please contact abanazar@inbox.ru
