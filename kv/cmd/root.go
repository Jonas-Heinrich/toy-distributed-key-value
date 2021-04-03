package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var release bool

var rootCmd = &cobra.Command{
	Use:   "kv",
	Short: "kv is a toy distributed key/value store",
	Long:  `kv is a toy distributed key/value store as a project for my Bachelor's Degree`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&release, "release", "r", false, "release mode")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
