package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log"
	"math/big"
)

type KeyExchange interface {
	GetKeyPair() ([]byte, []byte)
	GenerateSecret([]byte) []byte
}

type ecdhe struct {
	key *ecdsa.PrivateKey
}

// ECDHE return ECDH key exchange with specified curve
func ECDHE(curve elliptic.Curve) (KeyExchange, error) {
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Printf("fail while generating ecdsa key: %v", err)
		return nil, err
	}

	keyex := &ecdhe{key: privateKey}
	return keyex, nil
}

// FromKeyPair returns ECDHE key exchange from marshalled key
func FromKeyPair(curve elliptic.Curve, private, public []byte) KeyExchange {
	priv := new(ecdsa.PrivateKey)
	priv.Curve = curve
	priv.D = new(big.Int).SetBytes(private)
	x, y := elliptic.Unmarshal(curve, public)
	priv.X = x
	priv.Y = y

	return &ecdhe{key: priv}
}

// P256 returns ECDH key exchange with P-256
func P256() (KeyExchange, error) {
	return ECDHE(elliptic.P256())
}

// FromP256 returns ECDH key exchange with P-256 from marshalled key
func FromP256(private, public []byte) KeyExchange {
	return FromKeyPair(elliptic.P256(), private, public)
}

// GetKeyPair returns public and private key pair
func (e *ecdhe) GetKeyPair() (private []byte, public []byte) {
	private = e.key.D.Bytes()
	public = elliptic.Marshal(e.key.Curve, e.key.X, e.key.Y)
	return
}

// GenerateSecret generate secret from private key and other party public key
func (e *ecdhe) GenerateSecret(public []byte) []byte {
	x, y := elliptic.Unmarshal(e.key.Curve, public)
	sharedKey, _ := e.key.Curve.ScalarMult(x, y, e.key.D.Bytes())

	return sharedKey.Bytes()
}
