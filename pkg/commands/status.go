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
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "status reporting subcommand",
	Long:  `can be used to return the status of a running campaign.`,
	Run: func(cmd *cobra.Command, args []string) {
		statusGet(cmd, args)
	},
}

func init() {
	// todo: implement the command line argument handling here
}

// statusGet will request the current status of the given campaign from the orchestrator
// and will print the output to the terminal
func statusGet(cmd *cobra.Command, args []string) {
	// todo: implement the orchestrator/POST requests to handle accessing the campaign DB
	// also "render" the status on the CLI here
}
