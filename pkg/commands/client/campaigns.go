package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	flagUsernameFile     string
	flagPasswordFile     string
	flagNotBefore        string
	flagActiveWindow     time.Duration
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
	defaultNotBefore := time.Now().Format(time.RFC3339Nano)
	defaultActiveWindow, err := time.ParseDuration("4w")
	if err != nil {
		log.Fatalf("error parsing default active window: %s", err)
	}

	campaignCreateCmd.Flags().StringVarP(&flagUsernameFile, "userfile", "u", "", "file of usernames (newline separated)")
	campaignCreateCmd.MarkFlagRequired("userfile")

	campaignCreateCmd.Flags().StringVarP(&flagPasswordFile, "passfile", "p", "", "file of passwords (newline separated)")
	campaignCreateCmd.MarkFlagRequired("passfile")

	campaignCreateCmd.Flags().StringVarP(&flagNotBefore, "notbefore", "b", defaultNotBefore, "requests will not start before this time")

	campaignCreateCmd.Flags().DurationVarP(&flagActiveWindow, "window", "w", defaultActiveWindow, "a duration that this campaign will be active (ex: 4w)")

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
		log.Fatalf("error reading lines from user file: %s", err)
	}

	passwords, err := readLines(flagPasswordFile)
	if err != nil {
		log.Fatalf("error reading lines from password file: %s", err)
	}

	parsedNotBefore, err := time.Parse(time.RFC3339Nano, flagNotBefore)
	if err != nil {
		log.Fatalf("error parsing notBefore time: %s", err)
	}

	parsedNotAfter := parsedNotBefore.Add(flagActiveWindow)

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
		log.Fatalf("error during request creation: %s", err)
	}

	err = authenticator.Auth(req)
	if err != nil {
		log.Fatalf("error during authentication: %s", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("error sending request: %s", err)
	}

	log.Debug(resp)
	log.Info("successfully created campaign")
}
