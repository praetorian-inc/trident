package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	flagUsernameFile     string
	flagPasswordFile     string
	flagNotBefore        string
	flagNotAfter         string
	flagScheduleInterval int
	flagProvider         string
)

var campaignCreateCmd = &cobra.Command{
	Use:   "campaign",
	Short: "campaign management subcommand",
	Long:  `can be used to create and examine existing password spraying campaigns`,
	Run: func(cmd *cobra.Command, args []string) {
		campaignCreate(cmd, args)
	},
}

func init() {
	campaignCreateCmd.Flags().StringVarP(&flagUsernameFile, "userfile", "u", "", "file of usernames (newline separated)")
	campaignCreateCmd.MarkFlagRequired("userfile")

	campaignCreateCmd.Flags().StringVarP(&flagPasswordFile, "passfile", "p", "", "file of passwords (newline separated)")
	campaignCreateCmd.MarkFlagRequired("passfile")

	campaignCreateCmd.Flags().StringVarP(&flagNotBefore, "notbefore", "b", "", "requests will not start before this time")
	campaignCreateCmd.MarkFlagRequired("notbefore")

	campaignCreateCmd.Flags().StringVarP(&flagNotAfter, "notafter", "a", "", "requests will not occur after this time")
	campaignCreateCmd.MarkFlagRequired("notafter")

	campaignCreateCmd.Flags().IntVarP(&flagScheduleInterval, "interval", "i", 0, "requests will happen with this interval between them")
	campaignCreateCmd.MarkFlagRequired("interval")

	campaignCreateCmd.Flags().StringVarP(&flagProvider, "provider", "v", "", "this is the authentication platform you are attacking")
	campaignCreateCmd.MarkFlagRequired("provider")

	rootCmd.AddCommand(campaignCreateCmd)
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func campaignCreate(cmd *cobra.Command, args []string) {
	orchestrator := viper.GetString("orchestrator-url")
	providers := viper.GetStringMap("providers")

	users, err := readLines(flagUsernameFile)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	passwords, err := readLines(flagPasswordFile)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	parsedNotBefore, err := time.Parse(time.RFC3339Nano, flagNotBefore)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	parsedNotAfter, err := time.Parse(time.RFC3339Nano, flagNotAfter)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"not_before":        parsedNotBefore,
		"not_after":         parsedNotAfter,
		"schedule_interval": flagScheduleInterval,
		"users":             users,
		"passwords":         passwords,
		"provider":          flagProvider,
		"provider_metadata": providers[flagProvider],
	})

	req, err := http.NewRequest("POST", orchestrator+"/campaign", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	argoToken := creds.Token

	fmt.Printf("token: %s", argoToken)

	req.Header.Add("cf-access-token", argoToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	fmt.Printf("response: %v", resp)
}
