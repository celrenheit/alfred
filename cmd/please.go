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
	"os"
	"strings"

	"github.com/celrenheit/alfred/parser"
	"github.com/celrenheit/alfred/wallet"
	"github.com/manifoldco/promptui"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
)

// pleaseCmd represents the import command
var pleaseCmd = &cobra.Command{
	Use:     "please",
	Short:   "please",
	Aliases: []string{"p"},
	Long:    `send XLM to someone`,
	Example: "alfred please send 20 XLM from master to jennifer",
	PreRunE: middlewares(checkDB, checkSecret),
	Run: func(cmd *cobra.Command, args []string) {
		var query string
		switch {
		case len(args) == 1:
			query = args[0]
		case len(args) > 1:
			query = strings.Join(args, " ")
		}

		statement, err := parser.Parse(query)
		if err != nil {
			fatal(err)
		}

		switch req := statement.(type) {
		case *parser.SendRequest:
			err := sendRequest(cmd, req)
			if err != nil {
				fatalf(describeHorizonError(err))
			}
		default:
			fatalf("unsupported statement type: %T", statement.Kind())
		}
	},
}

func describeHorizonError(err error) string {
	if err == nil {
		return ""
	}

	e, ok := err.(*horizon.Error)
	if !ok {
		return err.Error()
	}
	pb := e.Problem
	return fmt.Sprintf("%s (%s)", pb.Title, string(pb.Extras["result_codes"]))
}

func init() {
	RootCmd.AddCommand(pleaseCmd)

	pleaseCmd.Flags().BoolP("yes", "y", false, "if set, no confirmation prompt will be shown")
	viper.BindPFlags(pleaseCmd.Flags())
}

func sendRequest(cmd *cobra.Command, req *parser.SendRequest) error {
	if req.Currency != "XLM" {
		return errors.New("only XLM is supported right now")
	}

	client := getClient(viper.GetBool("testnet"))

	path := viper.GetString("db")
	secret := viper.GetString("secret")
	m, err := wallet.OpenSecretString(path, secret)
	if err != nil {
		return err
	}

	var (
		from = req.From
		to   = req.To
	)

	var src *keypair.Full
	if from != "" {
		var w *wallet.Wallet
		if addr, err := keypair.Parse(from); err == nil {
			w = m.WalletByAddress(addr.Address())
		} else {
			w = m.WalletByName(from)
			if w == nil {
				return fmt.Errorf("wallet '%s' not found", from)
			}
		}

		src = w.Keypair.(*keypair.Full)
	} else {
		sel := promptui.Select{
			Label: "Select Wallet",
			Items: m.Stellar.Wallets,
		}

		idx, _, err := sel.Run()
		if err != nil {
			return err
		}

		src = m.Stellar.Wallets[idx].Keypair.(*keypair.Full)
	}

	var memo build.TransactionMutator
	if to != "" {
		if addr, err := keypair.Parse(to); err == nil { // to custom address
			to = addr.Address()
		} else {
			if w := m.WalletByName(to); w != nil { // between wallet
				to = w.Keypair.Address()
			} else if contact, ok := m.Stellar.Contacts[to]; ok { // to contact
				to = contact.Address
				if contact.Memo != nil {
					memo = contact.Memo.ToTransactionMutator()
				}
			} else {
				return fmt.Errorf("destination '%s' not found", to)
			}
		}
	} else {
		var toList []string
		for name, _ := range m.Stellar.Contacts {
			toList = append(toList, name)
		}
		prompt := promptui.SelectWithAdd{
			Label:    "Destination",
			AddLabel: "Another address",
			Items:    toList,
			Validate: func(input string) error {
				_, err := keypair.Parse(input)
				return err
			},
		}

		_, name, err := prompt.Run()
		if err != nil {
			return err
		}

		contact := m.Stellar.Contacts[name]
		to = contact.Address
		if contact.Memo != nil {
			memo = contact.Memo.ToTransactionMutator()
		}
	}

	_, exists, err := getAccount(client, src.Address())
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("source account does exists, please fund it first")
	}

	_, exists, err = getAccount(client, to)
	if err != nil {
		return err
	}

	var txnMutator build.TransactionMutator
	if exists {
		txnMutator = build.Payment(
			build.Destination{AddressOrSeed: to},
			build.NativeAmount{Amount: req.Amount},
		)
	} else {
		txnMutator = build.CreateAccount(
			build.Destination{AddressOrSeed: to},
			build.NativeAmount{Amount: req.Amount},
		)
	}

	opts := []build.TransactionMutator{
		build.SourceAccount{src.Seed()},
		build.TestNetwork,
		build.AutoSequence{SequenceProvider: client},
		txnMutator,
	}
	if memo != nil {
		opts = append(opts, memo)
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
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_RIGHT)
		table.Append([]string{"Amount", req.Amount})
		table.Append([]string{"Currency", req.Currency})
		table.Append([]string{"Source", src.Address()})
		table.Append([]string{"Destination", to})
		table.Render()

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
