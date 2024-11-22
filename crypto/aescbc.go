package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
)

type aesCbc struct {
	EncodedCipher

	key []byte
	iv  []byte

	block cipher.Block
}

var (
	errCipertextTooShort           = errors.New("ciper text to short")
	errInvalidIv                   = errors.New("iv to short must be 16")
	errInvalidInput                = errors.New("invalid input")
	errIvToloong                   = errors.New("iv to long")
	errChipherNotMultipleBlockSize = errors.New("cipherText is not a multiple of the block size")
)

func NewAESCBC(key, iv []byte) (Cipher, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	newIv, err := GenerateIV(iv)
	if err != nil {
		return nil, err
	}

	crpypt := aesCbc{
		key:           key,
		iv:            newIv,
		block:         block,
		EncodedCipher: EncodedCipher{},
	}

	crpypt.EncodedCipher.Cipher = &crpypt

	return &crpypt, nil
}

func (a *aesCbc) Encrypt(input []byte) ([]byte, error) {

	aaa := a.pKCS7Padding(input, aes.BlockSize)
	ciphertext := make([]byte, len(aaa))

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errChipherNotMultipleBlockSize
	}

	mode := cipher.NewCBCEncrypter(a.block, a.iv)
	mode.CryptBlocks(ciphertext, aaa)

	return ciphertext, nil
}

func (a *aesCbc) Decrypt(input []byte) ([]byte, error) {
	mode := cipher.NewCBCDecrypter(a.block, a.iv)
	if len(input)%aes.BlockSize != 0 {
		return nil, errChipherNotMultipleBlockSize
	}

	mode.CryptBlocks(input, input)
	unpadded := a.pKCS7UnPadding(input, aes.BlockSize)
	return unpadded, nil
}

func (a *aesCbc) pKCS7Padding(cipthertext []byte, blockSize int) []byte {
	bufLen := len(cipthertext)
	padLen := blockSize - bufLen%blockSize
	padded := make([]byte, bufLen+padLen)
	copy(padded, cipthertext)
	for i := 0; i < padLen; i++ {
		padded[bufLen+i] = byte(padLen)
	}
	return padded
}

func (a *aesCbc) pKCS7UnPadding(ciphertext []byte, blockSize int) []byte {
	padding := len(ciphertext) - int(ciphertext[len(ciphertext)-1])
	buf := make([]byte, padding)
	copy(buf, ciphertext[:padding])
	return buf
}

func (a *aesCbc) IV() []byte {
	return a.iv
}

func GenerateIV(iv []byte) ([]byte, error) {
	if len(iv) > aes.BlockSize {
		return nil, errIvToloong
	}

	if len(iv) == aes.BlockSize {
		return iv, nil
	}

	strIv := hex.EncodeToString(iv)

	// and blank byte to iv
	for i := len(iv); i < aes.BlockSize; i++ {
		strIv += "00"
	}

	return hex.DecodeString(strIv)

}
