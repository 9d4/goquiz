package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/94d/goquiz/config"
	"github.com/94d/goquiz/entity"
	"github.com/94d/goquiz/server"
	"github.com/94d/goquiz/util"
	"github.com/spf13/pflag"
)

var (
	flags   = pflag.NewFlagSet("goquiz", pflag.ContinueOnError)
	version = "0.2.0"
)

func init() {
	flags.StringP("address", "a", "[::]:8085", "Address to listen on")
	flags.BoolP("seed", "s", false, "Run seeder before start")
	flags.BoolP("version", "v", false, "Show version")
	if err := flags.Parse(os.Args); err != nil {
		util.Fatal(err)
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
	if s, _ := flags.GetBool("seed"); s || config.FirstTime() {
		entity.Open(config.FirstTime())
		if entity.AllowAutoSeed {
			entity.Seed()
		}
	}

	server.Start()
}
