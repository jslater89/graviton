package config

import (
	"flag"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	TestMode       bool   `mapstructure:"test_mode"`
	DemoData       bool   `mapstructure:"demo_data"`
	MongoAddress   string `mapstructure:"mongo_address"`
	DBName         string `mapstructure:"db_name"`
	GoogleClientID string `mapstructure:"google_client_id"`
	GoogleSecret   string `mapstructure:"google_secret"`
}

func (c Config) GetDBName() string {
	if c.TestMode {
		return c.DBName + "-test"
	}
	return c.DBName
}

var config Config

func GetConfig() Config {
	return config
}

func OverrideTest(t bool) {
	config.TestMode = true
}

func Load(configOverride *string) error {

	var configFile *string
	if flag.Lookup("test_mode") == nil {
		flag.Bool("test_mode", false, "run in test mode (see config.toml comments)")
		flag.Bool("demo_data", false, "ensure demo data exists (see config.toml comments)")
		flag.String("server_address", ":10000", "address to run the photo service on")
		flag.String("mongo_address", "localhost", "address of the database instance to connect to")
		flag.String("db_name", "graviton", "mongo db name to use")
		flag.String("google_client_id", "", "google client ID for oauth2")
		flag.String("google_secret", "", "google secret for oauth2")

		configFile = flag.String("config_file", "config.toml", "the config file to use")
		pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	}

	if configFlag := flag.Lookup("config_file"); configFlag != nil {
		stringValue := configFlag.Value.String()
		configFile = &stringValue
	}
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
