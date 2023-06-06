package cmd

import (
	"fmt"

	"github.com/94d/goquiz/config"
	"github.com/94d/goquiz/db"
	"github.com/94d/goquiz/server"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var rootCmd = &cobra.Command{
	Short: "Simple quiz!",
	Use:   "goquiz",
	Run: func(cmd *cobra.Command, args []string) {
		if s, _ := cmd.Flags().GetBool("seed"); s {
			db.Connect()
			db.Migrate()
			db.Seed()
		}

		server.Start()
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
	rootCmd.Flags().BoolP("seed", "s", false, "Run seeder before start")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
