package trustdraw

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
)

const (
	rsaBits       = 1024
	aesCipherSize = 16
	cardLength    = aes.BlockSize
)

// Deal shuffles a set of 'cards', writing the deal file to the given deck io.Writer.
// It will contain all the information needed for the players to draw cards as part
// of a turn-based game without needing any further trust.
func Deal(deck io.Writer, cards []string, dealerPrv ed25519.PrivateKey, playerPubs ...*rsa.PublicKey) error {
	for _, card := range cards {
		if len(card) > cardLength {
			return fmt.Errorf("card '%s' too long, must be %d bytes or fewer", card, cardLength)
		}
	}
	if len(playerPubs) < 2 {
		return fmt.Errorf("two or more player keys are needed")
	}
	for i, pub := range playerPubs {
		if pub.Size() < rsaBits/8 {
			return fmt.Errorf(
				"player %d's key is too small (%d bits), must be at least %d bits",
				i+1, pub.Size()*8, rsaBits)
		}
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

	if _, err := fmt.Fprintf(writer, "trustdraw/v%s\n", Version); err != nil {
		return fmt.Errorf("unable to write the deck to the deal file: %w", err)
	}

	for _, card := range deckData {
		if _, err := fmt.Fprintf(writer, "%s\n", base64.RawStdEncoding.EncodeToString(card)); err != nil {
			return fmt.Errorf("unable to write the deck to the deal file: %w", err)
		}
	}
	for i, player := range allPlayerData {
		if _, err := fmt.Fprintf(writer, "\n%s\n", base64.RawStdEncoding.EncodeToString(player)); err != nil {
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

// shuffle shuffles a slice in-place.
func shuffle[T interface{}](slice []T) {
	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}

// encryptCard encrypts a card using the given AES cipher block.
func encryptCard(card string, blk cipher.Block) []byte {
	encCard := make([]byte, cardLength)
	// Pad the card name with zero bytes, if necessary.
	plainText := append([]byte(card), make([]byte, aes.BlockSize-len(card))...)
	blk.Encrypt(encCard, plainText)
	return encCard
}

// generateCardKeys generates one AES key for each player, and the AES cipher block derived from the XOR of all of them.
func generateCardKeys(n int) ([][]byte, cipher.Block, error) {
	playerKeys := make([][]byte, n)
	for i := range playerKeys {
		key := make([]byte, aesCipherSize)
		if _, err := crand.Read(key); err != nil {
			return nil, nil, err
		}
		playerKeys[i] = key
	}

	blk, err := aes.NewCipher(xor(playerKeys...))
	if err != nil {
		return nil, nil, err
	}

	return playerKeys, blk, nil
}

// encryptCardKeys encrypts the given card keys for one player's eyes only, using the given RSA public key.
func encryptCardKeys(cardKeys [][]byte, pub *rsa.PublicKey) ([]byte, error) {
	plain := bytes.Join(cardKeys, nil)

	aesKey := make([]byte, aesCipherSize)
	if _, err := crand.Read(aesKey); err != nil {
		return nil, err
	}
	blk, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	cipherText := make([]byte, aes.BlockSize+len(plain))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(crand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(blk, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plain)

	asymKey, err := rsa.EncryptOAEP(sha256.New(), crand.Reader, pub, aesKey, nil)
	if err != nil {
		return nil, err
	}

	return append(asymKey, cipherText...), nil
}
