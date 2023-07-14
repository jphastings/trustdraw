package trustdraw

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

func xor(keys ...[]byte) []byte {
	fullKey := bytes.Clone(keys[0])
	for _, key := range keys[1:] {
		for i, b := range key {
			fullKey[i] = fullKey[i] ^ b
		}
	}
	return fullKey
}

// extractStanzas splits a deal file into its 4 stanzas, verifying that the declared
// version is one this code can read.
func extractStanzas(dealFile io.Reader) ([]string, error) {
	data, err := ioutil.ReadAll(dealFile)
	if err != nil {
		return nil, err
	}

	stanzas := strings.Split(string(data), "\n\n")
	if len(stanzas) != 4 {
		return nil, fmt.Errorf("deal file not valid")
	}

	if _, err := verifyVersion(stanzas[0]); err != nil {
		return nil, err
	}

	return stanzas, nil
}

// fromAllowKey converts an allowKey to a card ID and a card's AES key.
func fromAllowKey(allowKey string) (int, []byte, error) {
	akBytes, err := base64.RawStdEncoding.DecodeString(allowKey)
	if err != nil {
		return 0, nil, fmt.Errorf("invalid allowKey")
	}

	return int(binary.LittleEndian.Uint16(akBytes[0:2])), akBytes[2:], nil
}

// toAllowKey creates an allowKey from a card ID and the card's AES key.
func toAllowKey(cardID int, secretCard []byte) string {
	cardIDBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(cardIDBytes, uint16(cardID))

	return base64.RawStdEncoding.EncodeToString(append(cardIDBytes, secretCard...))
}

// allowKeysToCardKey combines the allowKeys shared by other players to re-create the card key needed to decrypt the indicated card.
func (d *Game) allowKeysToCardKey(allowKeys []string) (int, cipher.Block, error) {
	keys := make([][]byte, len(allowKeys))
	var cardID int
	for i, allowKey := range allowKeys {
		thisCardID, secretCard, err := fromAllowKey(allowKey)
		if err != nil {
			return 0, nil, err
		}

		if i == 0 {
			cardID = thisCardID
		} else if cardID != thisCardID {
			return 0, nil, fmt.Errorf("allowKeys are not for the same card")
		}

		keys[i] = secretCard
	}

	// Make cipher from all the keys XORed together with this user's key for this card.
	cardKey, err := aes.NewCipher(xor(append(keys, d.keys[cardID])...))
	if err != nil {
		return 0, nil, fmt.Errorf("internal error; could not re-create card key cipher")
	}

	return cardID, cardKey, nil
}

// decryptCard turns decrypts the referenced card with the given cardKey.
// TODO: Add an HMAC so I can know I've decrypted them properly
func (g *Game) decryptCard(cardID int, cardKey cipher.Block) (string, error) {
	card := make([]byte, aesCipherSize)
	cardKey.Decrypt(card, g.cards[cardID])

	return strings.Trim(string(card), "\x00"), nil
}
