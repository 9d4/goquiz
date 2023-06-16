package config

import (
	"log"

	"github.com/94d/goquiz/auth"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var V *viper.Viper

func init() {
	V = viper.New()
	V.OnConfigChange(onConfigChange)

	V.SetDefault("secret", auth.GenerateSecret())
	V.SetDefault("adminUsername", "admin")
	V.SetDefault("adminPassword", "admin===")
}

func InitConfig() {
	V.AddConfigPath(".")
	V.SetConfigName("goquiz")
	V.AutomaticEnv()
	err := V.ReadInConfig()
	for err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatal(err)
		}

		log.Println("Generating default config...")
		V.SetConfigType("yml")
		errwrite := V.SafeWriteConfig()
		if errwrite != nil {
			log.Fatal(errwrite)
		}

		err = V.ReadInConfig()
	}

	log.Printf("Config used %s\n", V.ConfigFileUsed())
}

func onConfigChange(in fsnotify.Event) {
}
