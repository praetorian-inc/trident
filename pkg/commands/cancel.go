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
	"github.com/praetorian-inc/trident/pkg/db"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
)

var cancelCommand = &cobra.Command{
	Use:   "cancel",
	Short: "cancel campaign execution",
	Long:  `can be used to halt a running campaign and stop all further spraying.`,
	Run: func(cmd *cobra.Command, args []string) {
		cancelPost(cmd, args)
	},
}

func init() {
	cancelCommand.Flags().StringVarP(&campaignID, "campaign", "c", "1",
		"the identifier of the campaign.")
	err := cancelCommand.MarkFlagRequired("campaign")
	if err != nil {
		log.Fatalf("issue during argument parsing: %s", err)
	}

	campaignCmd.AddCommand(cancelCommand)
}

// cancelPost will post the parameters that make up the given campaign
// and print the parameters to the CLI
func cancelPost(cmd *cobra.Command, args []string) {
	orchestrator := viper.GetString("orchestrator-url")

	q := db.Query{
		Filter: map[string]interface{}{
			"id": campaignID,
		},
	}

	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(q)

	if err != nil {
		log.Fatalf("error encoding cancel json request: %s", err)
	}

	req, err := http.NewRequest("POST", orchestrator+"/cancel", buf)

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
	if resp.StatusCode != 200 {
		log.Fatalf("error cancelling campaign from server: %d", resp.StatusCode)
	}
}
