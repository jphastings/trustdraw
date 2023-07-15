package trustdraw

import (
	"crypto/rsa"
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
func OpenGame(dealFile io.Reader, playerPrv *rsa.PrivateKey, state string) (*Game, error) {
	stanzas, err := extractStanzas(dealFile)
	if err != nil {
		return nil, err
	}
	cardCount := len(strings.Split(stanzas[1], "\n"))

	game := Game{
		Players: len(strings.Split(stanzas[2], "\n")),
		cards:   make([][]byte, cardCount),
	}
	if err := game.LoadState(state); err != nil {
		return nil, fmt.Errorf("could not load game state: %w", err)
	}

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

		game.keys, err = decryptCardKeys(keyBlock, playerPrv, cardCount)
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

// State produces a base64 encoded string that represents the current state of the game.
func (g *Game) State() string {
	state := make([]byte, len(g.state))
	for i, player := range g.state {
		state[i] = byte(player)
	}

	return base64.RawStdEncoding.EncodeToString(state)
}

// LoadState loads the game state from a string encoded with State().
func (g *Game) LoadState(states string) error {
	cardCount := len(g.cards)
	if cardCount == 0 {
		return fmt.Errorf("can't load state before the cards have been loaded")
	}
	g.state = make([]PlayerNumber, cardCount)

	state, err := base64.RawStdEncoding.DecodeString(states)
	if err != nil {
		return err
	}

	for i, playerByte := range state {
		if playerByte > byte(g.Players) {
			return fmt.Errorf("player %d is not in this game", playerByte)
		}
		g.state[i] = PlayerNumber(playerByte)
	}

	return nil
}
