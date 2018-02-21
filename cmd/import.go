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
	"log"
	"strings"

	"github.com/stellar/go/keypair"

	"github.com/celrenheit/alfred/wallet"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:     "import",
	Short:   "Import a wallet",
	Long:    `Import a wallet (should be private key)`,
	PreRunE: middlewares(checkDB, checkSecret),
	Run: func(cmd *cobra.Command, args []string) {
		validate := func(input string) error {
			if !strings.HasPrefix(input, "S") {
				return fmt.Errorf("this should be a private key")
			}
			return nil
		}

		prompt := promptui.Prompt{
			Label:    "What is the seed address ?",
			Validate: validate,
			Mask:     '*',
		}

		seed, err := prompt.Run()
		if err != nil {
			fatal(err)
			return
		}

		path := viper.GetString("db")
		secret := viper.GetString("secret")
		m, err := wallet.OpenSecretString(path, secret)
		if err != nil {
			fatal(err)
		}

		kp, err := keypair.Parse(seed)
		if err != nil {
			fatal(err)
		}

		var kpFull *keypair.Full
		switch v := kp.(type) {
		case *keypair.Full:
			kpFull = v
		default:
			log.Fatal("the key provided is not a seed")

		}

		name := cmd.Flag("name").Value.String()
		err = m.AddWallet(wallet.New(name, kpFull))
		if err != nil {
			fatal(err)
		}

		if err := wallet.Write(path, m); err != nil {
			fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(importCmd)

	importCmd.Flags().String("name", "", "name of the wallet")
	viper.BindPFlags(importCmd.Flags())
}
