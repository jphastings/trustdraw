package trustdraw

import (
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// PlayerNumber is 1-indexed (The first player is 1).
type PlayerNumber int

type Game struct {
	playerNumber PlayerNumber
	Players      int
	cards        [][]byte
	keys         [][]byte

	// state lists the player that each card has been given to.
	// 0 means the card is still in the deck to be drawn.
	state []PlayerNumber
}

// OpenGame opens a deal file, returning a Deal that can be used to draw cards.
// Make sure you have Verified the deck before using it.
func OpenGame(dealFile io.Reader, playerPrv *rsa.PrivateKey, state []PlayerNumber) (*Game, error) {
	stanzas, err := extractStanzas(dealFile)
	if err != nil {
		return nil, err
	}
	countCards := len(strings.Split(stanzas[1], "\n"))

	game := Game{
		Players: len(strings.Split(stanzas[2], "\n")),
		cards:   make([][]byte, countCards),
		state:   state,
	}
	game.state = append(game.state, make([]PlayerNumber, countCards-len(game.state))...)

	for i, card := range strings.Split(stanzas[1], "\n") {
		if game.cards[i], err = base64.RawStdEncoding.DecodeString(card); err != nil {
			return nil, fmt.Errorf("card %d is invalid", i+1)
		}
	}

	for i, playerData := range strings.Split(stanzas[2], "\n") {
		keyBlock, err := base64.RawStdEncoding.DecodeString(playerData)
		if err != nil {
			return nil, fmt.Errorf("player %d's data is invalid", i+1)
		}

		game.keys, err = decryptCardKeys(keyBlock, playerPrv)
		if err != nil {
			// Not our player block, try the next one
			continue
		}
		game.playerNumber = PlayerNumber(i + 1)
		break
	}
	if game.keys == nil {
		return nil, fmt.Errorf("the deal file wasn't made for the given player public key")
	}

	return &game, nil
}

// decryptCardKeys decrypts the given card key block with the given player's RSA public key.
func decryptCardKeys(playerData []byte, prv *rsa.PrivateKey) ([][]byte, error) {
	encAESKey := playerData[:rsaBits/8]
	cipherText := playerData[rsaBits/8:]

	aesKey, err := rsa.DecryptOAEP(sha256.New(), crand.Reader, prv, encAESKey, nil)
	if err != nil {
		return nil, err
	}

	blk, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	plainText := make([]byte, len(cipherText))
	stream := cipher.NewCTR(blk, cipherText[:aes.BlockSize])
	stream.XORKeyStream(plainText, cipherText[aes.BlockSize:])

	keys := make([][]byte, len(plainText)/aesCipherSize)

	var j int
	for i := 0; i < len(plainText); i += aesCipherSize {
		j += aesCipherSize
		// do what do you want to with the sub-slice, here just printing the sub-slices
		keys[i/aesCipherSize] = plainText[i:j]
	}

	return keys, nil
}
