package config

import (
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/brettpechiney/workout-service/config/param"
)

// Config is is a configuration implementation backed by Viper.
type Config struct {
	remote bool
	v      *viper.Viper
}

// Load returns a Config object that reads configuration settings.
func Load(configPaths []string) (*Config, error) {
	i := &Config{v: viper.New()}
	i.setDefaults()
	i.setupEnvVarReader()

	for _, dir := range configPaths {
		i.v.AddConfigPath(dir)
	}

	i.v.SetConfigName("application-properties")
	i.v.SetConfigType("toml")

	if err := i.v.ReadInConfig(); err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "unable to read configuration file")
		}
		log.Printf("no configuration file found; proceeding without one")
	}

	return i, nil
}

// Defaults returns a Config that has just the default values
// set. It will load neither local nor remote files.
func Defaults() *Config {
	i := &Config{v: viper.New()}
	i.setDefaults()
	return i
}

// Set overrides the configuration value. It is used for testing.
func (i *Config) Set(key string, value interface{}) {
	i.v.Set(key, value)
}

// DataSource returns the connection string of the database that
// stores Config application information.
func (i *Config) DataSource() string {
	return i.v.GetString(param.DataSource)
}

// LoggingLevel returns the application's logging level.
func (i *Config) LoggingLevel() string {
	return i.v.GetString(param.LoggingLevel)
}

func (i *Config) setDefaults() {
	const Source = "postgresql://maxroach@localhost:26257/workout?sslmode=disable"
	const Level = "INFO"
	i.v.SetDefault(param.DataSource, Source)
	i.v.SetDefault(param.LoggingLevel, Level)
}

func (i *Config) setupEnvVarReader() {
	i.v.AutomaticEnv()
	i.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}
