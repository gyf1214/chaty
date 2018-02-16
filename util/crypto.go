package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
)

func SHA256(data string) string {
	ret := sha256.Sum256([]byte(data))
	return base64.StdEncoding.EncodeToString(ret[:])
}

func RandomString(len int) (string, error) {
	buf := make([]byte, len)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", nil
	}
	return base64.StdEncoding.EncodeToString(buf), nil
}

func newCipher(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm, nil
}

func Encrypt(key []byte, data []byte) (string, string, error) {
	gcm, err := newCipher(key)
	if err != nil {
		return "", "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", err
	}

	crypted := gcm.Seal(nil, nonce, data, nil)
	out := base64.StdEncoding.EncodeToString(crypted)
	iv := base64.StdEncoding.EncodeToString(nonce)
	return out, iv, nil
}

func Decrypt(key []byte, data string, iv string) ([]byte, error) {
	gcm, err := newCipher(key)
	if err != nil {
		return nil, err
	}

	crypted, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	nonce, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nonce, crypted, nil)
}
