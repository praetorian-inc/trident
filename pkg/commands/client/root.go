package client

import (
	"net/url"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"trident/pkg/commands/auth"
)

var authenticator auth.Authenticator

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "trident-cli",
	Short: "command-line client for the trident password spraying system",
	Long: `used by an operator to input password spraying tasks into the
	orchestrator which will be then handed out to the registered dispatch
	nodes`,
}

func init() {
	// Use config file from the flag.
	viper.AddConfigPath("$HOME/.trident")
	viper.AddConfigPath("/etc/trident")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("error reading config: %s", err)
	}

	log.Infof("Using config file: %s", viper.ConfigFileUsed())

	url, err := url.Parse(viper.GetString("orchestrator-url"))
	if err != nil {
		log.Fatalf("error parsing orchestrator url: %s", err)
	}

	authenticator = &auth.ArgoAuthenticator{
		URL: url,
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("error during command execution: %s", err)
	}
}
