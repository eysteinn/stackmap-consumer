package prog

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

const (
	salt = "stackmap"
)

func calculateFileHash(filePath string) (string, []byte, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	// Create a new SHA-256 hash object
	sha256Hash := sha256.New()

	// 1 MB chunks
	chunksize := 1024 * 1024
	// Read the file in chunks and update the hash object
	buffer := make([]byte, chunksize)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", nil, err
		}
		sha256Hash.Write(buffer[:n])
	}

	sha256Hash.Write([]byte(salt))

	// Get the hexadecimal representation of the hash
	hashInBytes := sha256Hash.Sum(nil)
	hashString := hex.EncodeToString(hashInBytes)

	return hashString, hashInBytes, nil
}
