package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

var (
	ErrEmptyKey     = errors.New("key cannot be nil")
	ErrInvalidKey   = errors.New("invalid key size")
	ErrInvalidNonce = errors.New("nonce must be a 12 bytes")
)

type aesGcm struct {
	EncodedCipher

	key  []byte
	iv   []byte
	aead *cipher.AEAD
}

// NewAESGCM return cipher AES-GCM
func NewAESGCM(key []byte, iv []byte) (Cipher, error) {
	if key == nil {
		return nil, ErrEmptyKey
	}

	if kLen := len(key); kLen != 16 && kLen != 32 {
		return nil, ErrInvalidKey
	}

	if iv == nil || len(iv) != 12 {
		return nil, ErrInvalidNonce
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	i := aesGcm{EncodedCipher: EncodedCipher{}, key: key, iv: iv, aead: &aesgcm}
	i.EncodedCipher.Cipher = &i

	return &i, nil
}

func (a *aesGcm) Encrypt(input []byte) ([]byte, error) {
	if input == nil {
		return nil, nil
	}

	return (*a.aead).Seal(nil, a.iv, input, nil), nil
}

func (a *aesGcm) Decrypt(input []byte) ([]byte, error) {
	if input == nil {
		return nil, nil
	}

	out, err := (*a.aead).Open(nil, a.iv, input, nil)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (a *aesGcm) IV() []byte {
	return a.iv
}
