package crypto

import (
	"encoding/base64"
	"encoding/hex"
)

type Cipher interface {
	Encrypt(input []byte) ([]byte, error)
	Decrypt(input []byte) ([]byte, error)
	IV() []byte
	EncryptToHex(input []byte) (string, error)
	DecryptFromHex(input string) ([]byte, error)
	EncryptToBase64(input []byte) (string, error)
	DecryptFromBase64(input string) ([]byte, error)
	IVToHex() string
	IVToBase64() string
}

type EncodedCipher struct {
	Cipher
}

func (ec *EncodedCipher) EncryptToHex(input []byte) (string, error) {
	out, err := ec.Encrypt(input)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(out), nil
}

func (ec *EncodedCipher) DecryptFromHex(input string) ([]byte, error) {
	out, err := hex.DecodeString(input)
	if err != nil {
		return nil, err
	}

	return ec.Decrypt(out)
}

func (ec *EncodedCipher) EncryptToBase64(input []byte) (string, error) {
	out, err := ec.Encrypt(input)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(out), nil
}

func (ec *EncodedCipher) DecryptFromBase64(input string) ([]byte, error) {
	out, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, err
	}

	return ec.Decrypt(out)
}

func (ec *EncodedCipher) IVToHex() string {
	return hex.EncodeToString(ec.IV())
}

func (ec *EncodedCipher) IVToBase64() string {
	return base64.StdEncoding.EncodeToString(ec.IV())
}
