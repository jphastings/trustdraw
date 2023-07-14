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

// verifyDrawCmd represents the verifyDraw command
var verifyDrawCmd = &cobra.Command{
	Use:   "verify-draw dealFile playerPrivateKey drawnCard allowKey…",
	Short: "Verify another player's drawn card",
	Args:  cobra.MinimumNArgs(4),
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

		valid, err := game.VerifyDraw(args[2], args[3:]...)
		if err != nil {
			return err
		}

		if valid {
			fmt.Println("✅ This was a valid draw")
		} else {
			fmt.Println("❌ This was not a valid draw")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(verifyDrawCmd)
}
