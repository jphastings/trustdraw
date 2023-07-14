package decks

import (
	"embed"
	"fmt"
	"os"
	"strings"
)

//go:embed *.txt
var decks embed.FS

func Load(name string) ([]string, error) {
	if cards, ok := LoadInBuilt(name); ok {
		return cards, nil
	}

	data, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("cannot load deck %s: %v", name, err)
	}

	return strings.Split(string(data), "\n"), nil
}

func LoadInBuilt(name string) ([]string, bool) {
	data, err := decks.ReadFile(name + ".txt")
	if err != nil {
		return nil, false
	}
	return strings.Split(string(data), "\n"), true
}
