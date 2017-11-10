package config

import (
	"flag"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	TestMode       bool   `mapstructure:"test_mode"`
	MongoAddress   string `mapstructure:"mongo_address"`
	DBName         string `mapstructure:"db_name"`
	GoogleClientID string `mapstructure:"google_client_id"`
	GoogleSecret   string `mapstructure:"google_secret"`
}

var config Config

func GetConfig() Config {
	return config
}

func Load(configOverride *string) error {
	flag.Bool("test_mode", false, "run in test mode (see config.toml comments)")
	flag.String("server_address", ":10000", "address to run the photo service on")
	flag.String("mongo_address", "localhost", "address of the database instance to connect to")
	flag.String("db_name", "graviton", "mongo db name to use")
	flag.String("google_client_id", "", "google client ID for oauth2")
	flag.String("google_secret", "", "google secret for oauth2")

	configFile := flag.String("config_file", "config.toml", "the config file to use")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	if configOverride != nil {
		viper.SetConfigFile(*configOverride)
	} else {
		viper.SetConfigFile(*configFile)
	}

	viper.ReadInConfig()
	viper.BindPFlags(pflag.CommandLine)

	err := viper.Unmarshal(&config)

	return err
}
