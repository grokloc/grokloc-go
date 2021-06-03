// Package security provides crypto and hashing support
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"

	"github.com/matthewhartstonge/argon2"
)

// KeyLen is the encryption key length
const KeyLen = 32

// EncodedSHA256 returns the encoded (base16) sha256sums
func EncodedSHA256(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// MakeKey returns a 32-len byte
func MakeKey(s string) ([]byte, error) {
	v := EncodedSHA256(s)
	bs := []byte(v[:KeyLen])
	if len(bs) != KeyLen {
		return nil, errors.New("cannot construct key")
	}
	return bs, nil
}

// Encrypt returns the hex-encoded AES symmetric encryption of s with key
func Encrypt(s string, key []byte) (string, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(gcm.Seal(nonce, nonce, []byte(s), nil)), nil
}

// Decrypt reverses Encrypt
// e is the crypted+encoded string
func Decrypt(e string, key []byte) (string, error) {
	d, err := hex.DecodeString(e)
	if err != nil {
		return "", err
	}
	c, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(d) < nonceSize {
		return "", err
	}
	nonce, msg := d[:nonceSize], d[nonceSize:]
	bs, err := gcm.Open(nil, nonce, msg, nil)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

// DerivePassword performs a one-way hash on a password using argon2
func DerivePassword(password string, cfg argon2.Config) (string, error) {
	raw, err := cfg.Hash([]byte(password), nil)
	if err != nil {
		return "", err
	}
	return string(raw.Encode()), nil
}

// VerifyPassword returns true if guess is the same as the password forming derived
func VerifyPassword(guess, derived string) (bool, error) {
	return argon2.VerifyEncoded([]byte(guess), []byte(derived))
}
