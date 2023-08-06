package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/scrypt"
)

const minKeyLength = 32
const minPasswordLength = 5

var (
	ErrInvalidPassword             = errors.New("invalid password - cannot be used for encryption or decryption")
	ErrEmptyInputText              = errors.New("input text provided is empty")
	ErrSecretKeyInsufficientLength = fmt.Errorf("secret key must be at least %d bytes", minKeyLength)
	ErrCannotDecrypt               = errors.New("cannot decrypt the encrypted input with the password provided")
)

// deriveKey takes a password and salt, then derives a key suitable for use in encryption
func deriveKey(password, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, 32)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, err
		}
	}

	/*
		NOTE recommended to regularly evalate pushing these params (N, r, p) up... however doing so will wreck
		ability to decrypt existing ciphertexts... may require module versioning in future to get around this
	*/
	key, err := scrypt.Key(password, salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, nil, err
	}

	return key, salt, nil
}

func isValidPassword(password []byte) bool {
	return len(password) >= minPasswordLength
}

func validCryptInputs(text, password []byte) error {
	if !isValidPassword(password) {
		return ErrInvalidPassword
	}
	if len(text) == 0 {
		return ErrEmptyInputText
	}
	return nil
}

// Encrypt accepts (not-empty) input text and encrypts it using a (valid) password
func Encrypt(text, password []byte) ([]byte, error) {
	err := validCryptInputs(text, password)
	if err != nil {
		return nil, err
	}
	return encrypt(text, password)
}

// Decrypt accepts (not-empty) encrypted text and decrypts it using a (valid) password
func Decrypt(text, password []byte) ([]byte, error) {
	err := validCryptInputs(text, password)
	if err != nil {
		return nil, err
	}
	return decrypt(text, password)
}

func encrypt(text, password []byte) ([]byte, error) {
	secretKey, salt, err := deriveKey(password, nil)
	if err != nil {
		return nil, err
	}
	if len(secretKey) < 32 {
		return nil, ErrSecretKeyInsufficientLength

	}

	c, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}

	// gcm or Galois/Counter Mode, is a mode of operation
	// for symmetric key cryptographic block ciphers
	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	// nonce is a populated by a cryptographically secure random sequence
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	cipherText := gcm.Seal(nonce, nonce, text, nil)

	cipherText = append(cipherText, salt...)
	return cipherText, nil
}

func decrypt(encryptedData, password []byte) ([]byte, error) {
	salt, encryptedData := encryptedData[len(encryptedData)-32:], encryptedData[:len(encryptedData)-32]
	secretKey, _, err := deriveKey(password, salt)
	if err != nil {
		return nil, err

	}
	if len(secretKey) < 32 {
		return nil, ErrSecretKeyInsufficientLength

	}

	c, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}

	// gcm or Galois/Counter Mode, is a mode of operation
	// for symmetric key cryptographic block ciphers
	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce, cipherText := encryptedData[:gcm.NonceSize()], encryptedData[gcm.NonceSize():]
	//gcm can panic - handle this here
	var plaintext []byte
	defer func() ([]byte, error) {
		if r := recover(); r != nil {
			return nil, ErrCannotDecrypt
		}
		return plaintext, nil
	}()

	plaintext, err = gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, ErrCannotDecrypt
	}

	return plaintext, nil
}
