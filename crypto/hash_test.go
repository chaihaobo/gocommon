package crypto

import (
	"crypto/sha256"
	"testing"
)

func TestNewArgon2IDHash(t *testing.T) {
	hash := NewArgon2IDHash(&GeneratePwdParams{
		Memory:      256,
		Iterations:  2,
		Parallelism: 2,
		SaltLength:  8,
		KeyLength:   32,
	})
	if hash == nil {
		t.Fatal("constructor failed")
	}
}

func TestArgon2ID_Generate(t *testing.T) {
	hash := NewArgon2IDHash(&GeneratePwdParams{
		Memory:      256,
		Iterations:  2,
		Parallelism: 2,
		SaltLength:  8,
		KeyLength:   32,
	})
	encoded, _ := hash.Generate("very_easy_password")
	if encoded == "" {
		t.Fatal("invalid hash")
	}
}

func TestArgon2ID_Compare(t *testing.T) {
	hash := NewArgon2IDHash(&GeneratePwdParams{
		Memory:      256,
		Iterations:  2,
		Parallelism: 2,
		SaltLength:  8,
		KeyLength:   32,
	})
	password := "very_easy_password"
	encoded, _ := hash.Generate(password)
	if equal, _ := hash.Compare(password, encoded); !equal {
		t.Fatal("invalid hash")
	}
}

func TestHMAC(t *testing.T) {
	hash := HMAC(sha256.New, []byte("secret"), "signature")
	if hash == "" {
		t.Fatal("invalid hash")
	}
}
