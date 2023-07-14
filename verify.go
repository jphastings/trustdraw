package trustdraw

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

func VerifyDeal(dealFile io.Reader, dealerPub ed25519.PublicKey) (int, int, error) {
	stanzas, err := extractStanzas(dealFile)
	if err != nil {
		return 0, 0, err
	}

	cards, err := verifyCards(stanzas[1])
	if err != nil {
		return 0, 0, err
	}

	players, err := verifyPlayers(stanzas[2])
	if err != nil {
		return cards, 0, err
	}

	if err := verifySignature(stanzas, dealerPub); err != nil {
		return cards, players, err
	}

	return cards, players, nil
}

func verifyVersion(version string) (string, error) {
	parts := strings.Split(version, "/")
	if parts[0] != "TrustDraw" {
		return "", fmt.Errorf("not a deal file")
	}
	if parts[1] != "v1.0" {
		return "", fmt.Errorf("unknown deal file version: %s", version)
	}
	return parts[1], nil
}

func verifyCards(cards string) (int, error) {
	lines := strings.Split(cards, "\n")
	if len(lines) < 1 {
		return 0, fmt.Errorf("no cards in deal")
	}

	for i, line := range lines {
		key, err := base64.RawStdEncoding.DecodeString(line)
		if err != nil {
			return 0, fmt.Errorf("card %d is invalid", i+1)
		}
		if len(key) != 16 {
			return 0, fmt.Errorf("card %d is invalid", i+1)
		}
	}

	return len(lines), nil
}

func verifyPlayers(playerBlock string) (int, error) {
	players := strings.Split(playerBlock, "\n")
	for i, player := range players {
		if _, err := base64.RawStdEncoding.DecodeString(player); err != nil {
			return 0, fmt.Errorf("player %d's data is invalid", i+1)
		}
	}

	return len(players), nil
}

func verifySignature(stanzas []string, dealerPub ed25519.PublicKey) error {
	data := strings.Join(stanzas[0:3], "\n\n") + "\n"
	sig, err := base64.RawStdEncoding.DecodeString(stanzas[3])
	if err != nil {
		return fmt.Errorf("deal file signature is badly formed")
	}

	if !ed25519.Verify(dealerPub, []byte(data), sig) {
		return fmt.Errorf("deck was not shuffled by the specified dealer")
	}
	return nil
}
