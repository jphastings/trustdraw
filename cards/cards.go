package cards

import (
	"embed"
	"fmt"
	"os"
	"strings"
)

//go:embed *.txt
var cardsFS embed.FS

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
	data, err := cardsFS.ReadFile(name + ".txt")
	if err != nil {
		return nil, false
	}
	return strings.Split(string(data), "\n"), true
}
