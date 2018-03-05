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
	"path/filepath"

	"github.com/celrenheit/alfred/wallet"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// uploadCmd represents the import command
var uploadCmd = &cobra.Command{
	Use:     "upload",
	Short:   "upload",
	Aliases: []string{"p"},
	Long:    `send XLM to someone`,
	Example: "alfred upload ./text.txt",
	PreRunE: middlewares(checkDB, checkSecret),
	Run: func(cmd *cobra.Command, args []string) {
		path := viper.GetString("db")
		secret := viper.GetString("secret")
		m, err := wallet.OpenSecretString(path, secret)
		if err != nil {
			fatal(err)
		}

		if len(args) != 1 {
			fatal("only one file allowed")
		}

		fpath := args[0]
		data, err := ioutil.ReadFile(fpath)
		if err != nil {
			fatal(err)
		}

		name := filepath.Base(fpath)
		header := &Header{
			Filename: name,
			Length:   uint64(len(data)),
			Checksum: sha256sum(data),
		}

		parts := SplitIntoParts(name, data)

		kvs := make([]KVData, len(parts)+1)
		kvs[0] = header
		for i := 0; i < len(parts); i++ {
			kvs[i+1] = parts[i]
		}

		src, err := selectWallet(m)
		if err != nil {
			fatal(err)
		}

		err = submitData(viper.GetBool("testnet"), viper.GetBool("yes"), src, kvs)
		if err != nil {
			fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(uploadCmd)

	viper.BindPFlags(uploadCmd.Flags())
}
