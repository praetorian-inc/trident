package client

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"trident/pkg/commands/auth"
)

var cfgFile string

//TODO(dallas): see what up about making this an interface type
var creds = &auth.ArgoAuthenticator{}

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
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Using config file:", viper.ConfigFileUsed())

	url, err := url.Parse(viper.GetString("orchestrator-url"))
	if err != nil {
		fmt.Printf("error authenticating: %v", err)
		os.Exit(1)
	}

	creds = &auth.ArgoAuthenticator{
		Token: "",
		URL:   url,
	}

	if err := creds.Auth(); err != nil {
		fmt.Printf("error authenticating: %v", err)
		os.Exit(1)
	}
}

func Execute() {
	if err := creds.Auth(); err != nil {
		fmt.Printf("error authenticating: %v", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
