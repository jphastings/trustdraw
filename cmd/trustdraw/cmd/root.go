package cmd

import (
	"fmt"
	"os"

	"github.com/jphastings/trustdraw"
	"github.com/spf13/cobra"
)

var cmdVersion = 0

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Version: fmt.Sprintf("%s.%d", trustdraw.Version, cmdVersion),
	Use:     "trustdraw",
	Short:   "CLI tools for TrustDraw zero-trust card dealing",
	Long:    `Tooling for the TrustDraw protocol, for dealing and playing with a shuffled deck of cards in the open, using RSA, AES and Ed25519 encryption.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("state", "", "Path to the game state file to use")
}
