package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	flagReturnedFields string
	flagFilterFile     string
)

var resultsCmd = &cobra.Command{
	Use:   "results",
	Short: "results reporting subcommand",
	Long:  `can be used to return results from the server about the currently configured campaigns`,
	Run: func(cmd *cobra.Command, args []string) {
		resultsGet(cmd, args)
	},
}

func init() {
	resultsCmd.Flags().StringVarP(&flagReturnedFields, "return", "r", "*", "the list of fields you would like to see from the results (comma-separated string)")

	resultsCmd.Flags().StringVarP(&flagFilterFile, "filter", "f", "", "file containing your desired results filter")
	resultsCmd.MarkFlagRequired("passfile")

	rootCmd.AddCommand(resultsCmd)
}

func resultsGet(cmd *cobra.Command, args []string) {
	orchestrator := viper.GetString("orchestrator-url")

	fields := strings.Split(strings.ReplaceAll(flagReturnedFields, " ", ""), ",")

	filterFile, err := os.Open(flagFilterFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer filterFile.Close()

	byteValue, _ := ioutil.ReadAll(filterFile)

	var filter map[string]interface{}
	json.Unmarshal([]byte(byteValue), &filter)

	requestBody, err := json.Marshal(map[string]interface{}{
		"ReturnedFields": fields,
		"Filter":         filter,
	})

	req, err := http.NewRequest("POST", orchestrator+"/results", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	err = authenticator.Auth(req)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	fmt.Printf("response: %v", string(respBody))
}
