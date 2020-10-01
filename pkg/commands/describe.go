// Copyright 2020 Praetorian Security, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
//	"os"
//	"strings"

//	"github.com/jedib0t/go-pretty/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// identifier for the campaign
	campaignID string
)

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "campaign describe reporting subcommand",
	Long:  `can be used to return the parameters that makeup a given campaign.`,
	Run: func(cmd *cobra.Command, args []string) {
		describeGet(cmd, args)
	},
}

func init() {
	// todo: implement the command line argument handling here
	describeCmd.Flags().StringVarP(&campaignID, "campaign", "c", "*",
		"the identifier of the campaign.")

	rootCmd.AddCommand(describeCmd)
}

// describeGet will retrieve the parameters that make up the given campaign
// and print the parameters to the CLI
func describeGet(cmd *cobra.Command, args []string) {
	log.Infof("campaign id: %s", campaignID)
	// todo: implement the orchestrator/POST requests to handle accessing the campaign DB
	// also "render" the status on the CLI here
	orchestrator := viper.GetString("orchestrator-url")

	var filter = fmt.Sprintf("{ID:%s}", campaignID) 
	

	// build our request to the orchestrator.
	// return all fields (*) and the filter is the campaignID
	requestBody, err := json.Marshal(map[string]interface{}{
		"ReturnedFields": "*",
		"Filter":         filter,
	})

	log.Infof("request body: %s", requestBody)

	if err != nil {
		log.Fatalf("error during JSON marshalling for request body: %s", err)
	}

	req, err := http.NewRequest("POST", orchestrator+"/describe", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("error during request creation: %s", err)
	}

	// add Cloudflare Access token to our request
	err = authenticator.Auth(req)
	if err != nil {
		log.Fatalf("error during authentication: %s", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("error sending request: %s", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	// handle the results from the server
	
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}
	log.Infof("response: %s", respBody)
	var results []map[string]interface{}

	err = json.Unmarshal(respBody, &results)
	if err != nil {
		log.Fatalf("error parsing response json: %s", err)
	}

	//if flagOutputFormat == "json" {
	fmt.Print(string(respBody))
	//	return
	//}

	
}
