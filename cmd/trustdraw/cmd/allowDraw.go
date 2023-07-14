/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/jphastings/trustdraw"
	"github.com/jphastings/trustdraw/internal/cmdhelpers"
	"github.com/spf13/cobra"
)

// allowDrawCmd represents the allow-draw command
var allowDrawCmd = &cobra.Command{
	Use:   "allow-draw dealFile playerPrivateKey playerNumber",
	Short: "Allows a specified player to draw a card",
	Long:  `Retrieves the allowKey that can be shared with the other player(s) to allow them to draw a card.`,
	Args:  cobra.ExactArgs(3),
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

		intendedPlayer, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("player number must be an integer")
		}
		if intendedPlayer < 1 || intendedPlayer > game.Players {
			return fmt.Errorf("player #%d is not a part of this game (there are %d players)", intendedPlayer, game.Players)
		}

		allowKey, err := game.AllowDraw(trustdraw.PlayerNumber(intendedPlayer))
		if err != nil {
			return fmt.Errorf("could not allow card to be drawn: %w", err)
		}

		fmt.Print(allowKey)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(allowDrawCmd)
}
