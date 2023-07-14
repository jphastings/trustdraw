package cmdhelpers

import (
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func LoadDealerKey(path string) (ed25519.PrivateKey, error) {
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read dealer key (%s): %w", path, err)
	}

	pemBlock, _ := pem.Decode(pemBytes)
	if pemBlock == nil {
		return nil, fmt.Errorf("invalid dealer PEM file (%s)", path)
	}

	if pemBlock.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("dealer PEM file (%s) is not a private Ed25519 key", path)
	}

	key, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("dealer PEM file (%s) is not a private Ed25519 key", path)
	}

	edKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("dealer PEM file (%s) is not a private Ed25519 key", path)
	}

	return edKey, nil
}

func LoadPlayerKey(path string) (*rsa.PublicKey, error) {
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read player key (%s): %w", path, err)
	}

	pemBlock, _ := pem.Decode(pemBytes)
	if pemBlock == nil {
		return nil, fmt.Errorf("invalid player PEM file (%s)", path)
	}

	if pemBlock.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("player PEM file (%s) is not a public RSA key", path)
	}

	key, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("player PEM file (%s) is not a public RSA key", path)
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("player PEM file (%s) is not a public RSA key", path)
	}

	return rsaKey, nil
}
