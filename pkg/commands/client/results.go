package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	flagReturnedFields string
	flagFilter         string
	flagOutputFormat   string
)

var (
	DefaultReturnedFields = []string{
		"id",
		"username",
		"password",
		"valid",
	}
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
	resultsCmd.Flags().StringVarP(&flagReturnedFields, "return", "r", "*",
        "the list of fields you would like to see from the results (comma-separated string)")

	resultsCmd.Flags().StringVarP(&flagFilter, "filter", "f", `{"valid":true}`,
        "filter on db results (specified in JSON)")

	resultsCmd.Flags().StringVarP(&flagOutputFormat, "format", "o", "table",
        "output format (table, csv, json)")
	rootCmd.AddCommand(resultsCmd)
}

func resultsGet(cmd *cobra.Command, args []string) {
	orchestrator := viper.GetString("orchestrator-url")

	fields := strings.Split(strings.ReplaceAll(flagReturnedFields, " ", ""), ",")

	var filter map[string]interface{}
	json.Unmarshal([]byte(flagFilter), &filter)
	requestBody, err := json.Marshal(map[string]interface{}{
		"ReturnedFields": fields,
		"Filter":         filter,
	})

	req, err := http.NewRequest("POST", orchestrator+"/results", bytes.NewBuffer(requestBody))
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
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	var results []map[string]interface{}

	err = json.Unmarshal(respBody, &results)
	if err != nil {
		log.Fatalf("error parsing response json: %s", err)
	}

	if flagOutputFormat == "json" {
		fmt.Print(string(respBody))
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	if flagReturnedFields == "*" {
		fields = DefaultReturnedFields
	}

	header := make(table.Row, 0, len(fields))
	for _, field := range fields {
		header = append(header, field)
	}
	t.AppendHeader(header)

	for _, result := range results {
		var row table.Row
		for _, field := range fields {
			v := result[field]
			row = append(row, v)
		}
		t.AppendRows([]table.Row{row})
	}

	if flagOutputFormat == "csv" {
		t.RenderCSV()
		return
	}

	t.Render()
}
