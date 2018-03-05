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
	"io/ioutil"
	"sort"
	"strings"

	"github.com/celrenheit/alfred/wallet"
	"github.com/stellar/go/clients/horizon"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// downloadCmd represents the import command
var downloadCmd = &cobra.Command{
	Use:     "download",
	Short:   "download",
	Aliases: []string{"p"},
	Long:    `download file`,
	Example: "alfred download mywallet text.txt",
	PreRunE: middlewares(checkDB),
	Run: func(cmd *cobra.Command, args []string) {
		path := viper.GetString("db")
		secret := viper.GetString("secret")
		m, err := wallet.OpenSecretString(path, secret)
		if err != nil {
			fatal(err)
		}

		if len(args) != 2 {
			fatal("account and file expected")
		}

		accName, filename := args[0], args[1]

		kp := getAddress(m, accName)

		client := getClient(viper.GetBool("testnet"))
		acc, exists, err := getAccount(client, kp.Address())
		if err != nil {
			fatal(err)
		}
		if !exists {
			fatal("account does not exist")
		}

		header, parts, err := parseAccountData(acc, filename)
		if err != nil {
			fatal(err)
		}

		data, err := GetData(header, parts)
		if err != nil {
			fatal(err)
		}

		err = ioutil.WriteFile(filename, data, 0775)
		if err != nil {
			fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(downloadCmd)

	viper.BindPFlags(downloadCmd.Flags())
}

func parseAccountData(acc horizon.Account, filename string) (header *Header, parts []*Part, err error) {
	for key := range acc.Data {
		value, err := acc.GetData(key)
		if err != nil {
			return nil, nil, err
		}
		switch {
		case strings.HasPrefix(key, headerPrefix+filename):
			header, err = ParseHeader(key, value)
			if err != nil {
				return nil, nil, err
			}
		case strings.HasPrefix(key, filename):
			part, err := ParsePart(filename, key, value)
			if err != nil {
				return nil, nil, err
			}
			parts = append(parts, part)
		default:
			continue
		}
	}

	sort.Slice(parts, func(i, j int) bool {
		return parts[i].Offset < parts[j].Offset
	})

	return header, parts, nil
}
