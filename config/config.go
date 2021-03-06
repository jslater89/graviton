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
	SSLCert         string   `mapstructure:"sslCertPath"`
	SSLKey          string   `mapstructure:"sslKeyPath"`
	ServerAddress   string   `mapstructure:"serverAddress"`
	RedirectAddress string   `mapstructure:"redirectAddress"`
	CorsOrigins     []string `mapstructure:"corsOrigins"`
	MongoAddress    string   `mapstructure:"mongoAddress"`
	DBName          string   `mapstructure:"dbName"`
	GoogleClientID  string   `mapstructure:"googleClientId"`
	GoogleSecret    string   `mapstructure:"googleSecret"`
	ServerRedirect  string   `mapstructure:"serverRedirect"`
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
	if flag.Lookup("testMode") == nil {
		flag.Bool("testMode", false, "run in test mode (see config.toml comments)")
		flag.Bool("demoData", false, "ensure demo data exists (see config.toml comments)")
		flag.Bool("useSSL", false, "whether to use echo AutoTLS; requires sslCertPath and sslKeyPath")
		flag.String("sslCertPath", "/path/to/cert", "full path to SSL certificate")
		flag.String("sslKeyPath", "/path/to/key", "full path to SSL private key")
		flag.String("serverAddress", "localhost:10000", "address to run the graviton service on")
		flag.String("mongoAddress", "localhost", "address of the database instance to connect to")
		flag.String("dbName", "graviton", "mongo db name to use")
		flag.String("googleClientId", "", "google client ID for oauth2")
		flag.String("googleSecret", "", "google secret for oauth2")
		flag.String("redirectAddress", "http://localhost:8080/#/authenticated", "address to redirect to after oauth, to get Graviton bearer token")
		flag.String("serverRedirect", "http://localhost:10000", "external address to the server, for oauth redirects")

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
	} else if configFile != nil {
		viper.SetConfigFile(*configFile)
	} else {
		viper.SetConfigName("config")
	}

	viper.ReadInConfig()
	viper.BindPFlags(pflag.CommandLine)

	err := viper.Unmarshal(&config)

	return err
}
