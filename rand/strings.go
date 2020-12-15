package rand

import (
	"crypto/rand"
	"encoding/base64"
)

// RememberTokenBytes is the int for generating random tokens
const RememberTokenBytes = 32

// Byte helps generate random bytes using crypto/rand package
func Byte(n int) ([]byte, error) {
	bytes := make([]byte, n)
	n, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// String takes in an integer and converts to a random string 
func String(n int) (string, error) {
	bytes, err := Byte(n)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), err
}

// RememberToken uses String to generate random strings
func RememberToken() (string, error) {
	return String(RememberTokenBytes)
}