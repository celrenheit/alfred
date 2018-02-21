// Copyright © 2018 Salim Alami Idrissi
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
	"strconv"

	"github.com/celrenheit/alfred/parser"
	"github.com/manifoldco/promptui"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	alfredAddress = "GCDMBL2SDMM74I2EOM5XHF7LMMDXFEJQIZ5N2ORK6HBSHM5INLALFRED"
)

// donateCmd represents the import command
var donateCmd = &cobra.Command{
	Use:     "donate",
	Short:   "donate",
	Aliases: []string{"p"},
	Long:    `send XLM to someone`,
	Example: "alfred donate",
	PreRunE: middlewares(checkDB, checkSecret),
	Run: func(cmd *cobra.Command, args []string) {
		var amount string
		switch len(args) {
		case 1:
			amount = args[0]
		default:
			prompt := promptui.Prompt{
				Label: "Amount",
				Validate: func(input string) error {
					_, err := strconv.ParseFloat(input, 64)
					return err
				},
			}

			var err error
			amount, err = prompt.Run()
			if err != nil {
				fatal(err)
			}
		}

		err := sendRequest(cmd, &parser.SendRequest{
			Amount:   amount,
			Currency: "XLM",
			To:       alfredAddress,
		})
		if err != nil {
			fatalf(describeHorizonError(err))
		}
		fmt.Println("Thank You ♥️")
		fmt.Println("Keep on rockin' 🚀")
	},
}

func init() {
	RootCmd.AddCommand(donateCmd)

	viper.BindPFlags(donateCmd.Flags())
}
