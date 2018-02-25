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
	"errors"
	"fmt"

	"github.com/celrenheit/alfred/assets"
	"github.com/celrenheit/alfred/wallet"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stellar/go/build"
)

// trustCmd represents the import command
var trustCmd = &cobra.Command{
	Use:     "trust",
	Short:   "trust",
	Long:    `send XLM to someone`,
	Example: "alfred trust MOBI (GXXX)",
	PreRunE: middlewares(checkDB, checkSecret),
	Run: func(cmd *cobra.Command, args []string) {
		path := viper.GetString("db")
		secret := viper.GetString("secret")
		m, err := wallet.OpenSecretString(path, secret)
		if err != nil {
			fatal(err)
		}

		var asset *assets.Asset
		switch len(args) {
		case 0:
			fatal("no asset selected")
		case 1:
			asset, err = selectAsset(args[0])
			if err != nil {
				fatal(err)
			}
		case 2:
			asset = &assets.Asset{
				BuilderAsset: build.CreditAsset(args[0], args[1]),
			}
		}

		if err := trust(m, *asset); err != nil {
			fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(trustCmd)

	viper.BindPFlags(trustCmd.Flags())
}

func trust(m *wallet.Alfred, asset assets.Asset) error {
	client := getClient(viper.GetBool("testnet"))

	src, err := selectWallet(m)
	if err != nil {
		return err
	}

	acc, exists, err := getAccount(client, src.Address())
	if err != nil {
		return err
	}

	if !exists {
		return errors.New("account does not exists")
	}

	if hasTrustline(acc, asset) {
		return errors.New("account already has this trustline")
	}

	opts := []build.TransactionMutator{
		build.SourceAccount{src.Seed()},
		build.AutoSequence{SequenceProvider: client},
		build.Trust(asset.BuilderAsset.Code, asset.BuilderAsset.Issuer),
	}

	if viper.GetBool("testnet") {
		opts = append(opts, build.TestNetwork)
	} else {
		opts = append(opts, build.PublicNetwork)
	}

	tx, err := build.Transaction(opts...)
	if err != nil {
		return err
	}

	txe, err := tx.Sign(src.Seed())
	if err != nil {
		return err
	}

	txeB64, err := txe.Base64()
	if err != nil {
		return err
	}

	if !viper.GetBool("yes") {
		_, err = (&promptui.Prompt{
			Label:     "Are you sure",
			IsConfirm: true,
		}).Run()
		if err != nil {
			return err
		}
	}

	resp, err := client.SubmitTransaction(txeB64)
	if err != nil {
		return err
	}

	fmt.Println(resp.Hash)
	return nil
}
