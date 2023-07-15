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

		stateFile := cmdhelpers.StateFile(cmd.Flag("state").Value.String(), args[0], args[1])
		state, err := cmdhelpers.ReadOrMake(stateFile)
		if err != nil {
			return fmt.Errorf("the statefile was not writeable: %w", err)
		}

		game, err := trustdraw.OpenGame(deal, playerPrv, state)
		if err != nil {
			return err
		}

		card, allowKey, alreadyDrawn, err := game.Draw(args[2:]...)
		if err != nil {
			return err
		}

		if err := os.WriteFile(stateFile, []byte(game.State()), 0600); err != nil {
			return fmt.Errorf("could not save game state: %w", err)
		}

		verb := "have drawn"
		if alreadyDrawn {
			verb = "previously drew"
		}
		fmt.Printf("You %s: %s\nProve with: %s\n", verb, card, allowKey)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(drawCmd)
}
