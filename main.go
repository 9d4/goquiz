package main

import (
	"log"
	"os"

	"github.com/94d/goquiz/config"
	"github.com/94d/goquiz/entity"
	"github.com/94d/goquiz/server"
	"github.com/spf13/pflag"
)

var flags = pflag.NewFlagSet("goquiz", pflag.ContinueOnError)

func init() {
	flags.StringP("address", "a", "[::]:8085", "Address to listen on")
	flags.BoolP("seed", "s", false, "Run seeder before start")
	if err := flags.Parse(os.Args); err != nil {
		log.Fatal(err)
	}

	config.V.BindPFlag("address", flags.ShorthandLookup("a"))
	config.InitConfig()
}

func main() {
	if s, _ := flags.GetBool("seed"); s {
		entity.Open()
		entity.Seed()
	}

	server.Start()
}
