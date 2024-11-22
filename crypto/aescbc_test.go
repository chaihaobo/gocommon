package crypto

import (
	"encoding/hex"
	"fmt"
	"log"
	"testing"

	"github.com/bmizerany/assert"
)

func TestAesCbcHex(t *testing.T) {
	key, _ := hex.DecodeString("6368616e676520746869732070617373")
	iv, _ := hex.DecodeString("31363139323435323532383838000000")

	cbc, err := NewAESCBC(key, iv)
	if err != nil {
		log.Println(err)
		return
	}

	plaintext := "exampleplaintext"
	encrypted, err := cbc.Encrypt([]byte(plaintext))
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := cbc.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}

	str := string(decrypted)
	fmt.Printf("result: %s\n", string(decrypted))

	if str != plaintext {
		t.Fatal("different plaintext & decrypted")
	}
}

func TestAesCbcBase64(t *testing.T) {
	key, _ := hex.DecodeString("3437363634376561313139333664636664333063316435363961616130633337")
	iv, _ := hex.DecodeString("31363139323737353632343634313233")
	fmt.Println("key ", key)
	fmt.Println("iv ", iv)

	cbc, err := NewAESCBC(key, iv)
	if err != nil {
		log.Println(err)
		return
	}

	plaintext := "exampleplaintext a nytama aku arya anjimm bgt"
	encrypted, err := cbc.EncryptToBase64([]byte(plaintext))
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("encrypted ", encrypted)

	decrypted, err := cbc.DecryptFromBase64(encrypted)
	if err != nil {
		t.Fatal(err)
	}

	str := string(decrypted)
	fmt.Printf("result: %s\n", str)

	if str != plaintext {
		t.Fatal("different plaintext & decrypted")
	}
}

func TestIv(t *testing.T) {
	iv, err := hex.DecodeString("31363139323435323532383838000000")
	assert.Equal(t, nil, err)

	failedIv, err := hex.DecodeString("31363139323435323532383838")
	assert.Equal(t, nil, err)

	ivT, err := GenerateIV(failedIv)
	assert.Equal(t, nil, err)
	assert.Equal(t, iv, ivT)
}
