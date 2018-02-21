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
	"os"

	"github.com/olekukonko/tablewriter"

	"github.com/celrenheit/alfred/wallet"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// balancesCmd represents the balances command
var balancesCmd = &cobra.Command{
	Use:     "balances",
	Short:   "Display balances",
	Long:    `Display balances`,
	PreRunE: middlewares(checkDB),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := viper.GetString("db")
		secret := viper.GetString("secret")
		m, err := wallet.OpenSecretString(path, secret)
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		header := []string{"Wallet", "Currency", "Balance"}

		var rows [][]string
		for _, w := range m.Stellar.Wallets {
			balances, err := w.Balances(viper.GetBool("testnet"))
			if err != nil {
				row := []string{w.Name, "error", err.Error()}
				rows = append(rows, row)
				continue
			}

			for _, b := range balances {
				code := b.Asset.Code
				if b.Asset.Type == "native" {
					code = "XLM"
				}
				row := []string{w.Name, code, b.Balance}
				rows = append(rows, row)
			}
		}

		table.SetRowLine(true)
		table.SetAutoMergeCells(true)
		table.SetHeader(header)
		table.AppendBulk(rows)
		table.Render()

		return nil
	},
}

func init() {
	RootCmd.AddCommand(balancesCmd)

	viper.BindPFlags(balancesCmd.Flags())
}
