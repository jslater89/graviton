package config

import (
	"flag"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	TestMode        bool     `mapstructure:"testMode"`
	DemoData        bool     `mapstructure:"demoData"`
	UseSSL          bool     `mapstructure:"useSSL"`
	SSLCache        string   `mapstructure:"sslCache"`
	ServerAddress   string   `mapstructure:"serverAddress"`
	RedirectAddress string   `mapstructure:"redirectAddress"`
	CorsOrigins     []string `mapstructure:"corsOrigins"`
	MongoAddress    string   `mapstructure:"mongoAddress"`
	DBName          string   `mapstructure:"dbName"`
	GoogleClientID  string   `mapstructure:"googleClientId"`
	GoogleSecret    string   `mapstructure:"googleSecret"`
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
		flag.Bool("testMode", false, "run in test mode (see config.toml comments)")
		flag.Bool("demoData", false, "ensure demo data exists (see config.toml comments)")
		flag.Bool("useSSL", false, "whether to use echo AutoTLS")
		flag.String("sslCache", "/var/www/.cache", "where to store SSL certs")
		flag.String("serverAddress", "localhost:10000", "address to run the graviton service on")
		flag.String("mongoAddress", "localhost", "address of the database instance to connect to")
		flag.String("dbName", "graviton", "mongo db name to use")
		flag.String("googleClientId", "", "google client ID for oauth2")
		flag.String("googleSecret", "", "google secret for oauth2")
		flag.String("redirectAddress", "http://localhost:8080/#/authenticated", "redirect address to receive bearer after oauth")

		configFile = flag.String("configFile", "config.toml", "the config file to use")
		pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	}

	if configFlag := flag.Lookup("configFile"); configFlag != nil {
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
