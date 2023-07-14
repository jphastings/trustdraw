package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"os"

	"github.com/jphastings/trustdraw"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	cards := []string{"A♥️", "2♣️", "3♦️", "4♠️"}

	// Generate a new Ed25519 keypair for the dealer.
	_, dealerPrv, err := ed25519.GenerateKey(rand.Reader)
	check(err)

	// Generate two RSA key pairs for the players.
	player1Prv, err := rsa.GenerateKey(rand.Reader, 1024)
	check(err)
	player2Prv, err := rsa.GenerateKey(rand.Reader, 1024)
	check(err)

	// Deal the cards, writing the deal file to stdout
	check(trustdraw.Deal(os.Stdout, cards, dealerPrv, &player1Prv.PublicKey, &player2Prv.PublicKey))
}
