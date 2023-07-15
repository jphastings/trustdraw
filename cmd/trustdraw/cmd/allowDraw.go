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

		stateFile := cmdhelpers.StateFile(cmd.Flag("state").Value.String(), args[0], args[1])
		state, err := cmdhelpers.ReadOrMake(stateFile)
		if err != nil {
			return fmt.Errorf("the statefile was not writeable: %w", err)
		}

		game, err := trustdraw.OpenGame(deal, playerPrv, state)
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
		if err == trustdraw.ErrNoCardsLeft {
			_, _ = fmt.Fprintf(os.Stderr, "❌ There are no cards left to draw\n")
		} else if err != nil {
			return fmt.Errorf("could not get allowKey: %w", err)
		}

		if err := os.WriteFile(stateFile, []byte(game.State()), 0600); err != nil {
			return fmt.Errorf("could not save game state: %w", err)
		}

		fmt.Print(allowKey)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(allowDrawCmd)
}
