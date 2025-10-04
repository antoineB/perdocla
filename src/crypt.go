package src

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
)

func generateKey(keyFilename string) ([]byte, error) {
	key := make([]byte, 32)
	length, _ := rand.Read(key)
	if length != 32 {
		return []byte{}, fmt.Errorf("Random generator didn't generate a long enough key")
	}

	err := os.WriteFile(keyFilename, []byte(hex.EncodeToString(key)), 0400)
	if err != nil {
		return []byte{}, err
	}

	return key, nil
}

func encryptFile(filename string, key []byte) ([]byte, error) {
    fileContent, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, aesGCM.NonceSize())
    if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }

    ciphertext := aesGCM.Seal(nonce, nonce, fileContent, nil)
    return ciphertext, nil
}

func decryptFile(filename string, key []byte) ([]byte, error) {
    ciphertext, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

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
        return nil, errors.New("ciphertext too short")
    }

    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
    fileContent, err := aesGCM.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, err
    }

    return fileContent, nil
}

func readKeyFile(filename string) ([]byte, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return []byte{}, err
	}

	if len(content) == 32 {
		return content, nil
	}

	byteContent, err := hex.DecodeString(string(content))

	if err != nil {
		return []byte{}, err
	}

	if len(byteContent) != 32 {
		return []byte{}, fmt.Errorf("The file %s should be 32 bytes length", filename)
	}

	return byteContent, nil
}