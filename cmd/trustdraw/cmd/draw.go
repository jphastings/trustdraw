/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/jphastings/trustdraw"
	"github.com/jphastings/trustdraw/internal/cmdhelpers"
	"github.com/spf13/cobra"
)

// drawCmd represents the draw command
var drawCmd = &cobra.Command{
	Use:   "draw dealFile playerPrivateKey allowKey…",
	Short: "Draws a card from the dealt deck.",
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		deal, err := os.Open(args[0])
		if err != nil {
			return err
		}

		playerPrv, err := cmdhelpers.LoadPlayerPrivateKey(args[1])
		if err != nil {
			return err
		}

		game, err := trustdraw.OpenGame(deal, playerPrv, []trustdraw.PlayerNumber{1, 2, 1, 2})
		if err != nil {
			return err
		}

		card, allowKey, err := game.Draw(args[2:]...)
		if err != nil {
			return err
		}

		fmt.Printf("You drew: %s\nProve with: %s\n", card, allowKey)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(drawCmd)
}
