package crypto

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestAesGcmHex(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	iv, _ := hex.DecodeString("bb8ef84243d2ee95a41c6c57")

	c, err := NewAESGCM(key, iv)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("IV: %s\n", c.IVToHex())

	plaintext := "exampleplaintext"
	encrypted, err := c.Encrypt([]byte(plaintext))
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("result: %x\n", encrypted)

	decrypted, err := c.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}

	str := string(decrypted)
	fmt.Printf("result: %s\n", string(decrypted))

	if str != plaintext {
		t.Fatal("different plaintext & decrypted")
	}
}

func TestAesGcmBase64(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	iv, _ := hex.DecodeString("bb8ef84243d2ee95a41c6c57")

	c, err := NewAESGCM(key, iv)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("IV: %s\n", c.IVToBase64())

	plaintext := "exampleplaintext"
	encrypted, err := c.EncryptToBase64([]byte(plaintext))
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("result: %s\n", encrypted)

	decrypted, err := c.DecryptFromBase64(encrypted)
	if err != nil {
		t.Fatal(err)
	}

	str := string(decrypted)
	fmt.Printf("result: %s\n", string(decrypted))

	if str != plaintext {
		t.Fatal("different plaintext & decrypted")
	}
}
