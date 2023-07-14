/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"crypto/rsa"
	"fmt"
	"os"
	"path"

	"github.com/jphastings/trustdraw"
	decks "github.com/jphastings/trustdraw/cards"
	"github.com/jphastings/trustdraw/internal/cmdhelpers"
	"github.com/spf13/cobra"
)

// dealCmd represents the deal command
var dealCmd = &cobra.Command{
	Use:   "deal deck dealerPrivateKey playerPublicKey playerPublicKey…",
	Short: "Produce a Deal file for the specified players",
	Long:  `Produces a Deal file that holds all the information needed to hold a trustless game of cards for the players whose public keys afre provided.`,
	Args:  cobra.MinimumNArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		cards, err := decks.Load(args[0])
		if err != nil {
			return err
		}

		dealerPrv, err := cmdhelpers.LoadDealerPrivateKey(args[1])
		if err != nil {
			return err
		}

		playerPubs := make([]*rsa.PublicKey, len(args)-2)
		for i, arg := range args[2:] {
			playerPub, err := cmdhelpers.LoadPlayerPublicKey(arg)
			if err != nil {
				return err
			}
			playerPubs[i] = playerPub
		}

		if err := trustdraw.Deal(os.Stdout, cards, dealerPrv, playerPubs...); err != nil {
			return err
		}

		_, _ = fmt.Fprintf(os.Stderr, "\nDeal file with %d shuffled cards written to stdout\n", len(cards))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dealCmd)

	dealCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Fprintf(os.Stderr, `Usage: %s %s

%s

<deck>      One of the in-build decks (see below), or a path to a text file
            containing a list of 'card' names. They cannot be longer than 16
            bytes, and must be one per line (\n).

<dealerKey> The path to an Ed25519 private key in PEM format, used for signing
            the deck file. Generate a new Ed25519 key pair with:
              $ openssl genpkey -algorithm ed25519 > dealer.pem

<playerKey> Each must be the path to an RSA public key, at least 1024 bits
            long, in PEM format. Two or more player keys can be specified.
            Generate a private key with:
              $ openssl genpkey -algorithm rsa > playerX.pem
            And extract the public key with:
              $ openssl rsa -in playerX.pem -pubout -out playerX.pub.pem

The dealer must publish their public key for the players to trust the deck:
  $ openssl pkey -in dealer.pem -pubout -out dealer.pub.pem

In-build decks:

  standard52-fr A French-suited standard 52 card deck of cards: 🃑 🃒 🃓 etc…
  scrabble-en   An English Scrabble 100 tile set: 12×E 9×A 9×I 8×O etc…
  scrabble-es   A Spanish Scrabble 100 tile set: 12×A 1×CH 1×Ñ etc…
  escarbar      A Latin-American Scrabble 108 tile set: 12×E 3×LL 3×Ñ etc…`,
			path.Base(os.Args[0]), cmd.Use, cmd.Long)

		return nil
	})
}
