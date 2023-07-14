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

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify dealFile dealerPublicKey",
	Short: "Verifies a TrustDraw deal file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		deal, err := os.Open(args[0])
		if err != nil {
			return err
		}

		key, err := cmdhelpers.LoadDealerPublicKey(args[1])
		if err != nil {
			return err
		}

		cards, players, err := trustdraw.VerifyDeal(deal, key)
		if err != nil {
			return fmt.Errorf("%s is not a valid deal file: %v", args[0], err)
		}

		_, _ = fmt.Fprintf(os.Stderr, "%s is a valid deck of %d cards for %d players\n", args[0], cards, players)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}
