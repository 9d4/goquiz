package config

import (
	"fmt"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var V *viper.Viper

func init() {
	V = viper.New()
	V.OnConfigChange(onConfigChange)
}

func InitConfig() {
	V.AddConfigPath(".")
	V.SetConfigName("goquiz")
	V.AutomaticEnv()
	err := V.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatal(err)
		}

		log.Println("Generating default config...")
		V.SetConfigType("yml")
		errwrite := V.SafeWriteConfig()
		if errwrite != nil {
			log.Fatal(errwrite)
		}

		fmt.Print("\nNow you can configure the config file then run GoQuiz again\n")
		os.Exit(0)
	}

	log.Printf("Config used %s\n", V.ConfigFileUsed())
}

func onConfigChange(in fsnotify.Event) {
}
