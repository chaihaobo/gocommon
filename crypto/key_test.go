package crypto

import (
	"bytes"
	"crypto/elliptic"
	"log"
	"testing"
)

func TestECDHE(t *testing.T) {
	keyex, _ := ECDHE(elliptic.P521())
	if keyex == nil {
		t.Fatal("invalid key exchange")
	}
}

func TestFromKeyPair(t *testing.T) {
	key1, _ := P256()
	priv1, pub1 := key1.GetKeyPair()
	key2 := FromP256(priv1, pub1)
	priv2, pub2 := key2.GetKeyPair()
	if !bytes.Equal(priv1, priv2) || !bytes.Equal(pub1, pub2) {
		t.Fatal("invalid from key pair")
	}
}

func TestP256(t *testing.T) {
	keyex, _ := P256()
	if keyex == nil {
		t.Fatal("invalid key exchange")
	}
}

func TestEcdhe_GetKeyPair(t *testing.T) {
	keyex, _ := P256()
	priv, pub := keyex.GetKeyPair()
	if priv == nil || pub == nil {
		t.Fatal("failed to generate key pair")
	}
}

func TestEcdhe_GenerateSecret(t *testing.T) {
	k1, _ := P256()
	k2, _ := P256()
	_, pub1 := k1.GetKeyPair()
	_, pub2 := k2.GetKeyPair()

	secret1 := k1.GenerateSecret(pub2)
	secret2 := k2.GenerateSecret(pub1)

	log.Printf("%x", secret1)
	log.Printf("%x", secret2)

	if !bytes.Equal(secret1, secret2) {
		t.Fatal("failed to generate shared secret key")
	}
}
