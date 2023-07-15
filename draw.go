package trustdraw

import (
	"errors"
	"fmt"
)

var ErrNoCardsLeft = errors.New("no cards left to draw")

// AllowDraw retrieves the allowKey for that will allow the specified player to draw a card.
// An allowKey contains 2 bytes of card ID, followed by 16 bytes of the card's AES key.
func (g *Game) AllowDraw(intended PlayerNumber) (string, error) {
	if intended < 1 || intended > PlayerNumber(g.Players) {
		return "", fmt.Errorf("player %d is not in this game", intended)
	}

	for cardID, secretCard := range g.keys {
		if g.state[cardID] != 0 {
			continue
		}

		g.state[cardID] = intended
		return toAllowKey(cardID, secretCard), nil
	}

	return "", ErrNoCardsLeft
}

// Draw uses the allowKeys shared by other players to draw the relevant card.
func (g *Game) Draw(allowKeys ...string) (card string, allowKey string, alreadyDrawn bool, error error) {
	if len(allowKeys) != g.Players-1 {
		return "", "", false, fmt.Errorf("wrong number of allowKeys (%d needed, %d given)", g.Players, len(allowKeys))
	}
	cardID, cardKey, err := g.allowKeysToCardKey(allowKeys)
	if err != nil {
		return "", "", false, fmt.Errorf("could not re-create card key: %w", err)
	}

	card, err = g.decryptCard(cardID, cardKey)
	if err != nil {
		return "", "", false, fmt.Errorf("could not decrypt card: %w", err)
	}

	alreadyDrawn = g.state[cardID] != 0
	g.state[cardID] = g.playerNumber

	return card, toAllowKey(cardID, g.keys[cardID]), alreadyDrawn, nil
}

func (g *Game) VerifyDraw(testCard string, allowKeys ...string) (bool, error) {
	cardID, blk, err := g.allowKeysToCardKey(allowKeys)
	if err != nil {
		return false, fmt.Errorf("could not re-create card key: %w", err)
	}

	realCard, err := g.decryptCard(cardID, blk)
	if err != nil {
		return false, fmt.Errorf("could not decrypt card: %w", err)
	}

	return testCard == realCard, nil
}
