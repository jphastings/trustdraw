package trustdraw

import (
	"crypto/ed25519"
	"crypto/rsa"
	"io"
)

type PlayerNumber int

type Deck struct {
	playerKeys    [][]byte
	shuffledCards [][]byte

	CardUse []PlayerNumber
}

func OpenDeck(dealFile io.Reader, playerPrv *rsa.PrivateKey, dealerPub ed25519.PublicKey) (*Deck, error) {
	return nil, nil
}

func (d *Deck) Draw() (string, error) {
	return "", nil
}
