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
	// we want to support config directories in home or etc
	viper.AddConfigPath("$HOME/.trident")
	viper.AddConfigPath("/etc/trident")

	// config file name is config.yaml
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("error reading config: %s", err)
	}

	log.Infof("Using config file: %s", viper.ConfigFileUsed())

	// parse out the orchestrator server URL
	url, err := url.Parse(viper.GetString("orchestrator-url"))
	if err != nil {
		log.Fatalf("error parsing orchestrator url: %s", err)
	}

	// create the global authenticator that will be used to add an auth
	// token to each command that needs it
	authenticator = &auth.ArgoAuthenticator{
		URL: url,
	}
}

// Execute is the entrypoint into the cmd line interface. It will execute the
// desired subcommand and check for an error, reporting it if so
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("error during command execution: %s", err)
	}
}
