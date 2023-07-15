package trustdraw

import "crypto/aes"

const (
	rsaBits       = 1024
	aesCipherSize = 16
	cardLength    = aes.BlockSize
	maxCards      = 65536
	// Chosen so the largest player number fits into 1 base64 encoded byte, with player 0 being reserved
	maxPlayers = 191
)
