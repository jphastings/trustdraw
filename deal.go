package trustdraw

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"io"
)

// Deal shuffles a set of 'cards', writing the deal file to the given deck io.Writer.
// It will contain all the information needed for the players to draw cards as part
// of a turn-based game without needing any further trust.
func Deal(deck io.Writer, cards []string, dealerPrv ed25519.PrivateKey, playerPubs ...*rsa.PublicKey) error {
	if err := validateDealArgs(cards, playerPubs); err != nil {
		return err
	}

	deckData := make([][]byte, len(cards))
	allPlayerData := make([][]byte, len(playerPubs))
	players := len(playerPubs)
	allCardKeys := make([][][]byte, players)
	shuffle(cards)

	for i, card := range cards {
		cardKeys, blk, err := generateCardKeys(players)
		if err != nil {
			return fmt.Errorf("unable to shuffle the deck: %w", err)
		}

		deckData[i] = encryptCard(card, blk)
		for p, key := range cardKeys {
			if i == 0 {
				allCardKeys[p] = make([][]byte, len(cards))
			}
			allCardKeys[p][i] = key
		}
	}

	for p, cardKeys := range allCardKeys {
		playerData, err := encryptCardKeys(cardKeys, playerPubs[p])
		if err != nil {
			return fmt.Errorf("unable to encrypt the card keys for player %d: %w", p, err)
		}
		allPlayerData[p] = playerData
	}

	var sigBytes bytes.Buffer
	writer := io.MultiWriter(&sigBytes, deck)

	if _, err := fmt.Fprintf(writer, "TrustDraw/v%s\n\n", Version); err != nil {
		return fmt.Errorf("unable to write the deck to the deal file: %w", err)
	}

	for _, card := range deckData {
		if _, err := fmt.Fprintf(writer, "%s\n", base64.RawStdEncoding.EncodeToString(card)); err != nil {
			return fmt.Errorf("unable to write the deck to the deal file: %w", err)
		}
	}

	if _, err := fmt.Fprintln(writer); err != nil {
		return fmt.Errorf("unable to write the deck to the deal file: %w", err)
	}

	for i, player := range allPlayerData {
		if _, err := fmt.Fprintf(writer, "%s\n", base64.RawStdEncoding.EncodeToString(player)); err != nil {
			return fmt.Errorf("unable to write player %d's keys to the deal file: %w", i, err)
		}
	}

	// Write the signature to the deck file
	sig := ed25519.Sign(dealerPrv, sigBytes.Bytes())
	if _, err := fmt.Fprintf(deck, "\n%s", base64.RawStdEncoding.EncodeToString(sig)); err != nil {
		return fmt.Errorf("unable to write the signature to the deck file: %w", err)
	}

	return nil
}

func validateDealArgs(cards []string, playerPubs []*rsa.PublicKey) error {
	if len(cards) > maxCards {
		return fmt.Errorf("too many cards, max is %d", maxCards)
	}
	for _, card := range cards {
		if len(card) > cardLength {
			return fmt.Errorf("card '%s' too long, must be %d bytes or fewer", card, cardLength)
		}
	}
	if len(playerPubs) < 2 {
		return fmt.Errorf("two or more player keys are needed")
	}
	if len(playerPubs) > maxPlayers {
		return fmt.Errorf("no more than %d players are allowed", maxPlayers)
	}

	for i, pub := range playerPubs {
		if pub.Size() < rsaBits/8 {
			return fmt.Errorf(
				"player %d's key is too small (%d bits), must be at least %d bits",
				i+1, pub.Size()*8, rsaBits)
		}
	}

	return nil
}
