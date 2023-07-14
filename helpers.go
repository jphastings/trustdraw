package trustdraw

import "bytes"

func xor(keys ...[]byte) []byte {
	fullKey := bytes.Clone(keys[0])
	for _, key := range keys[1:] {
		for i, b := range key {
			fullKey[i] = fullKey[i] ^ b
		}
	}
	return fullKey
}
