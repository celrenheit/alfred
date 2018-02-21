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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stellar/go/keypair"
)

// fundCmd represents the import command
var fundCmd = &cobra.Command{
	Use:     "fund",
	Short:   "fund",
	Aliases: []string{"p"},
	Long:    `send XLM to someone`,
	Example: "alfred fund GXX",
	PreRunE: middlewares(checkDB),
	Run: func(cmd *cobra.Command, args []string) {
		if !viper.GetBool("testnet") {
			fatal("not yet supported on main net, stay tuned")
		}

		if len(args) != 1 {
			fatal("one argument is expected, either an address or the name of the wallet")
		}

		arg := args[0]

		addr, err := keypair.Parse(arg)
		if err != nil {
			fatal(err)
		}

		friendbotFund(addr.Address())
	},
}

func init() {
	RootCmd.AddCommand(fundCmd)

	viper.BindPFlags(fundCmd.Flags())
}
