package trustdraw

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"strings"
)

func xor(keys ...[]byte) []byte {
	fullKey := make([]byte, len(keys[0]))
	copy(fullKey, keys[0])
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
	data, err := io.ReadAll(dealFile)
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

// decryptCardKeys decrypts the given card key block with the given player's RSA public key.
func decryptCardKeys(playerData []byte, prv *rsa.PrivateKey, cardCount int) ([][]byte, error) {
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

	return keys[0:cardCount], nil
}

// shuffle shuffles a slice in-place.
func shuffle(slice []string) {
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
