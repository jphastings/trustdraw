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
	if _, _, err := VerifyDeal(dealFile, dealerPub); err != nil {
		return nil, err
	}
	return nil, nil
}

func (d *Deck) Draw() (string, error) {
	return "", nil
}
