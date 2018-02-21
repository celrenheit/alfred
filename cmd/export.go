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
	"encoding/csv"
	"log"
	"os"

	"github.com/stellar/go/keypair"

	"github.com/celrenheit/alfred/wallet"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// exportCmd represents the wallets command
var exportCmd = &cobra.Command{
	Use:     "export",
	Short:   "Export wallets in plaintext csv",
	Long:    `Export wallets in plaintext csv`,
	PreRunE: middlewares(checkDB, checkSecret),
	Run: func(cmd *cobra.Command, args []string) {
		path := viper.GetString("db")
		secret := viper.GetString("secret")
		m, err := wallet.OpenSecretString(path, secret)
		if err != nil {
			fatal(err)
		}

		if !m.IsUnlocked() {
			log.Fatal("you need to unlocked your wallet")
		}

		w := csv.NewWriter(os.Stdout)
		rows := make([][]string, 0)
		rows = append(rows, []string{"name", "address", "seed"})
		for _, w := range m.Stellar.Wallets {
			switch kp := w.Keypair.(type) {
			case (*keypair.FromAddress):
				log.Fatal("keypair is not unlocked")
			case (*keypair.Full):
				row := []string{w.Name, kp.Address(), kp.Seed()}
				rows = append(rows, row)
			}
		}

		if err := w.WriteAll(rows); err != nil {
			fatal(err)
		}

		w.Flush()

		if err := w.Error(); err != nil {
			fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(exportCmd)
}
