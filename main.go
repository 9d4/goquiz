package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/94d/goquiz/config"
	"github.com/94d/goquiz/entity"
	"github.com/94d/goquiz/server"
	"github.com/spf13/pflag"
)

var (
	flags   = pflag.NewFlagSet("goquiz", pflag.ContinueOnError)
	version = "0.1.0"
)

func init() {
	flags.StringP("address", "a", "[::]:8085", "Address to listen on")
	flags.BoolP("seed", "s", false, "Run seeder before start")
	flags.BoolP("version", "v", false, "Show version")
	if err := flags.Parse(os.Args); err != nil {
		log.Fatal(err)
	}

	if v, _ := flags.GetBool("version"); v {
		platform := runtime.GOOS
		arch := runtime.GOARCH

		fmt.Printf("GoQuiz Version %s (Platform: %s, Arch: %s)\n", version, platform, arch)
		os.Exit(0)
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
