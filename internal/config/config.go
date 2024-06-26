package config

import (
	"net/url"

	"github.com/spf13/viper"
)

type Config struct {
	// Address to bind the webserver to.
	BindAddr string

	// Configuration options for the database connection.
	Database DatabaseConfig

	// Boolean indicating if the application should be run in development mode.
	//
	// Development mode enables features such as per-request loading of templates or disabling
	// caching of static files to aid in development tasks. These features are less performant than
	// the default production-oriented configuration.
	DevMode bool

	// Set to true if the app will be served over HTTP
	Insecure bool

	GoogleClientID     string
	GoogleClientSecret string

	OAuthCallbackURL string

	RunMigrations bool
}

type DatabaseConfig struct {
	// The user to connect to the database as.
	User string

	// Password for the database user.
	Password string

	// Hostname of the database server.
	Hostname string

	// Port to connect to the database over.
	Port string

	// Name of the database to use.
	Name string
}

// Get the connection URL for the database.
func (c *DatabaseConfig) ConnectionURL() url.URL {
	var host string
	if c.Port == "" {
		host = c.Hostname
	} else {
		host = c.Hostname + ":" + c.Port
	}

	return url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   host,
		Path:   c.Name,
	}
}

func init() {
	viper.SetDefault("bind-address", ":8000")

	viper.BindEnv("database.host", "POSTGRES_HOSTNAME")
	viper.BindEnv("database.name", "POSTGRES_DB")
	viper.BindEnv("database.password", "POSTGRES_PASSWORD")
	viper.BindEnv("database.user", "POSTGRES_USER")

	viper.BindEnv("dev-mode", "DEV_MODE")
	viper.SetDefault("dev-mode", false)

	viper.BindEnv("google-client-id", "GOOGLE_CLIENT_ID")
	viper.BindEnv("google-client-secret", "GOOGLE_CLIENT_SECRET")

	viper.BindEnv("oauth-callback-url", "OAUTH_CALLBACK_URL")

	viper.SetDefault("run-migrations", false)
}

func FromEnvironment() (Config, error) {
	config := Config{
		BindAddr: viper.GetString("bind-address"),
		Database: DatabaseConfig{
			User:     viper.GetString("database.user"),
			Password: viper.GetString("database.password"),
			Hostname: viper.GetString("database.host"),
			Port:     viper.GetString("database.port"),
			Name:     viper.GetString("database.name"),
		},
		DevMode:            viper.GetBool("dev-mode"),
		GoogleClientID:     viper.GetString("google-client-id"),
		GoogleClientSecret: viper.GetString("google-client-secret"),
		OAuthCallbackURL:   viper.GetString("oauth-callback-url"),
		RunMigrations:      viper.GetBool("run-migrations"),
	}

	return config, nil
}
