package utils
import (

	"crypto/sha256"
	"encoding/hex"
	"io"


)


// Assuming you have a utility function like this
// CalculateChecksum calculates the SHA-256 checksum of the given file

func CalculateChecksum(file io.Reader) (string, error) {
    hasher := sha256.New()
    if _, err := io.Copy(hasher, file); err != nil {
        return "", err
    }
    return hex.EncodeToString(hasher.Sum(nil)), nil
}