package cryptobox

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"strings"
	"unicode"
)

// RSAGenerateKeysAsPemStr generate RSA
func RSAGenerateKeysAsPemStr() (privKey, pubKey string, err error) {
	RSAPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return privKey, pubKey, fmt.Errorf("rsa.GenerateKey failed: %v", err)
	}
	RSAPrivateKeyPEM := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(RSAPrivateKey),
	}

	RSAPublicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&RSAPrivateKey.PublicKey),
	}

	b := bytes.NewBuffer(nil)
	if err = pem.Encode(b, RSAPrivateKeyPEM); err != nil {
		return privKey, pubKey, fmt.Errorf("pem.Encode(privateKey) failed: %v", err)
	}
	privKey = b.String()
	b.Reset()
	if err = pem.Encode(b, RSAPublicKeyPEM); err != nil {
		return privKey, pubKey, fmt.Errorf("pem.Encode(publicKey) failed: %v", err)
	}
	pubKey = b.String()
	return
}

// RSAGetMagicKey return application/magic-public-key representation of the pubKey
func RSAGetMagicKey(pubKeyPem string) (magicKey string, err error) {
	block, _ := pem.Decode([]byte(pubKeyPem))
	if block == nil || block.Type != "PUBLIC KEY" {
		return magicKey, fmt.Errorf("%s is not a valid pem encoded public key", pubKeyPem)
	}
	pubKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return magicKey, fmt.Errorf("x509.ParsePKCS1PublicKey(block) failed: %v", err)
	}
	n := base64.RawURLEncoding.EncodeToString(pubKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pubKey.E)).Bytes())
	magicKey = "RSA." + n + "." + e
	return
}

// from thx https://github.com/emersion/go-ostatus (MIT)
func decodeString(s string) ([]byte, error) {
	s = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)

	// The spec says to use URL encoding without padding, but some implementations
	// add padding (e.g. Mastodon).
	if b, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return base64.URLEncoding.DecodeString(s)
}

// RSAParseMagicKey parse magic-public-key and return a rsa.PublicKey
func RSAParseMagicKey(magicKey string) (*rsa.PublicKey, error) {
	parts := strings.Split(magicKey, ".")
	if strings.ToUpper(parts[0]) != "RSA" {
		return nil, fmt.Errorf("%s is not an RSA magic-public-key", magicKey)
	}

	log.Printf("PArts %d", len(parts))

	if len(parts) != 3 {
		return nil, fmt.Errorf("%s is not a valid magiv-key", magicKey)
	}

	n, err := decodeString(parts[1])
	if err != nil {
		return nil, err
	}
	e, err := decodeString(parts[2])
	if err != nil {
		return nil, err
	}

	return &rsa.PublicKey{
		N: big.NewInt(0).SetBytes(n),
		E: int(big.NewInt(0).SetBytes(e).Int64()),
	}, nil

}
