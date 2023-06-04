package cmd

import (
	"fmt"

	"github.com/94d/goquiz/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var rootCmd = &cobra.Command{
	Short: "Simple quiz!",
	Use:   "goquiz",
	Run: func(cmd *cobra.Command, args []string) {
		ch := make(chan int, 1)
		<-ch
	},
}

var serverFlags = &pflag.FlagSet{}

func initFlags() {
	serverFlags.StringP("address", "a", "127.0.0.1:8080", "Address to listen on")
}

func init() {
	initFlags()

	config.V.BindPFlags(serverFlags)
	config.InitConfig()
	rootCmd.Flags().AddFlagSet(serverFlags)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
