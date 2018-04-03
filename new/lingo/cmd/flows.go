// Copyright © 2018 CODELINGO LTD hello@codelingo.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"github.com/juju/errors"

	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
)

// flowsCmd represents the flows command
var flowsCmd = &cobra.Command{
	Use:   "flows",
	Short: "List Flows",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if err := listFlows(cmd, args); err != nil {
			fmt.Fprint(os.Stderr, err.Error())
		}
	},
}

func init() {
	listCmd.AddCommand(flowsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// flowsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// flowsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	flowsCmd.Flags().StringP("owner", "o", "", "List all flows of the given owner")
	flowsCmd.Flags().StringP("name", "n", "", "Describe the named flow")
	flowsCmd.Flags().BoolP("intalled", "i", false, "List Flows installed in current project")

}

func listFlows(cmd *cobra.Command, args []string) error {

	owner := cmd.Flag("owner").Value.String()
	name := cmd.Flag("name").Value.String()

	baseFlowURL := baseDiscoveryURL + "flows"
	url := baseFlowURL + "/lingo_tenets.yaml"
	switch {
	case name != "":

		if owner == "" {
			return errors.New("owner flag must be set")
		}

		url = fmt.Sprintf("%s/%s/%s/lingo_flow.yaml",
			baseFlowURL, owner, name)

	case owner != "":
		url = fmt.Sprintf("%s/%s/lingo_owner.yaml",
			baseFlowURL, owner)
	}
	resp, err := http.Get(url)
	if err != nil {
		return errors.Trace(err)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Trace(err)
	}

	fmt.Println(string(data))
	return nil
}
