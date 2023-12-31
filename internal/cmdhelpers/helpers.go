package cmdhelpers

import (
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path"
	"strings"
)

func LoadDealerPrivateKey(path string) (ed25519.PrivateKey, error) {
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

func LoadDealerPublicKey(path string) (ed25519.PublicKey, error) {
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read dealer key (%s): %w", path, err)
	}

	pemBlock, _ := pem.Decode(pemBytes)
	if pemBlock == nil {
		return nil, fmt.Errorf("invalid dealer PEM file (%s)", path)
	}

	if pemBlock.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("dealer PEM file (%s) is not a public Ed25519 key", path)
	}

	key, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("dealer PEM file (%s) is not a public Ed25519 key", path)
	}

	edKey, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("dealer PEM file (%s) is not a public Ed25519 key", path)
	}

	return edKey, nil
}

func LoadPlayerPrivateKey(path string) (*rsa.PrivateKey, error) {
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read player key (%s): %w", path, err)
	}

	pemBlock, _ := pem.Decode(pemBytes)
	if pemBlock == nil {
		return nil, fmt.Errorf("invalid player PEM file (%s)", path)
	}

	if pemBlock.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("player PEM file (%s) is not a private RSA key", path)
	}

	key, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("player PEM file (%s) is not a private RSA key", path)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("player PEM file (%s) is not a private RSA key", path)
	}

	return rsaKey, nil
}

func LoadPlayerPublicKey(path string) (*rsa.PublicKey, error) {
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

// StateFile returns a default name for the state file of a game/player combination
func StateFile(explicit, dealFilePath, playerKeyPath string) string {
	if explicit != "" {
		return explicit
	}

	deal := strings.SplitN(path.Base(dealFilePath), ".", 2)
	player := strings.SplitN(path.Base(playerKeyPath), ".", 2)

	return fmt.Sprintf("%s.%s.state", deal[0], player[0])
}

func ReadOrMake(path string) (string, bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return "", false, err
		}
		return "", true, file.Close()
	} else if err != nil {
		return "", false, err
	}

	if info.Mode().Perm()&0200 == 0 {
		return "", false, fmt.Errorf("the path (%s) is not writeable", path)
	}

	data, err := os.ReadFile(path)
	return string(data), false, err
}
