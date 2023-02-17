package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

// FYI: https://go.dev/src/crypto/cipher/example_test.go

func Encrypt(key []byte, msg string) (string, error) {
	base64ed := base64.StdEncoding.EncodeToString([]byte(msg))
	plaintext := []byte(base64ed)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// prepand the nonce to the beginning of the ciphertext
	ciphertext := aesgcm.Seal(nonce, nonce, plaintext, nil)

	return fmt.Sprintf("%x", ciphertext), err
}

func Decrypt(key []byte, cipherText string) (string, error) {
	cipherTextBytes, err := hex.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesgcm.NonceSize()
	if len(cipherTextBytes) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// extract the nonce from the encrypted data
	nonce, encrypted := cipherTextBytes[:nonceSize], cipherTextBytes[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", err
	}

	// decode base64
	decoded, err := base64.StdEncoding.DecodeString(string(plaintext))
	if err != nil {
		return "", err
	}

	return string(decoded), err
}

func GenerateSecret() (string, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	hashedBytes := sha256.Sum256(randomBytes)
	return fmt.Sprintf("BSK-%s", base64.StdEncoding.EncodeToString(hashedBytes[:])), nil
}
